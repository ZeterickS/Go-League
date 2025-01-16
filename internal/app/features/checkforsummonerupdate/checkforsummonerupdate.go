package checkforsummonerupdate

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	apiHelper "discord-bot/internal/app/helper/api"
	"discord-bot/internal/app/helper/cdragon"
	databaseHelper "discord-bot/internal/app/helper/database"
	"discord-bot/internal/app/utility/gametoimage"
	"discord-bot/types/embed"
	"discord-bot/types/rank"
	"discord-bot/types/summoner"

	"github.com/bwmarrin/discordgo"
)

var discordSession *discordgo.Session
var channelID string

// Initialize sets the Discord session and channel ID for sending messages
func Initialize(session *discordgo.Session, chID string) {
	discordSession = session
	channelID = chID
}

func checkAndSendRankUpdate(discordSession *discordgo.Session, channelID string, currentSummoner, summoner *summoner.Summoner, rankType string) error {
	var currentRank, previousRank rank.Rank
	var pretttyRank string = ""
	switch rankType {
	case "Solo":
		currentRank = currentSummoner.SoloRank
		previousRank = summoner.SoloRank
		pretttyRank = "Solo/Duo"
	case "Flex":
		currentRank = currentSummoner.FlexRank
		previousRank = summoner.FlexRank
		pretttyRank = "Flex"
	default:
		log.Printf("Unknown rank type: %v", rankType)
		return fmt.Errorf("unknown rank type: %v", rankType)
	}

	if currentRank != previousRank {
		rankChange := currentRank - previousRank
		rankChangeString := ""
		if rankChange < 0 {
			rankChangeString = fmt.Sprintf("%v", rankChange)
		} else {
			rankChangeString = fmt.Sprintf("+%v", rankChange)
		}
		color := 0x00ff00 // Green color for LP gain
		if rankChange < 0 {
			color = 0xff0000 // Red color for LP loss
		}

		rankTier := strings.Split(currentRank.ToString(), " ")[0]
		rankTier = strings.ToLower(rankTier)
		fmt.Println(rankTier)

		// Get the ranked picture URL
		rankTierURL := cdragon.GetRankedPictureURL(rankTier)

		lastmatchid, err := apiHelper.GetLastRankedMatchIDbyPUUID(currentSummoner.PUUID)
		log.Printf("Last ranked match ID: %v", lastmatchid)

		if err != nil {
			log.Printf("Failed to fetch last ranked match ID: %v", err)
		}

		lastMatch, err := apiHelper.GetMatchByID(lastmatchid)
		log.Printf("Last match: %v", lastMatch)

		if err != nil {
			log.Printf("Failed to fetch last match: %v", err)
		}

		for i := 0; i < len(lastMatch.Teams); i++ {

			for _, participant := range lastMatch.Teams[i].Participants {
				log.Printf("Checking participant: %v", participant)
				log.Printf("Participant PUUID: %v, Current Summoner PUUID: %v\n", participant.Summoner.PUUID, currentSummoner.PUUID)
				if participant.Summoner.PUUID == currentSummoner.PUUID {

					log.Printf("Generating game image for participant: %v", participant)

					lastgameimage, err := gametoimage.GameToImage(participant)
					if err != nil {
						log.Printf("Failed to generate game image: %v", err)
						continue
					}

					// Use the URL directly
					embedmessage := embed.NewEmbed().
						SetAuthor(currentSummoner.GetNameTag(), cdragon.GetProfileIconURL(currentSummoner.ProfileIconID), fmt.Sprintf("https://www.op.gg/summoners/euw/%v-%v", currentSummoner.Name, currentSummoner.TagLine)).
						SetTitle(fmt.Sprintf("%v-Rank Update | %v LP", pretttyRank, rankChangeString)).
						AddField("Solo/Duo-Rank", currentSummoner.SoloRank.ToString()).
						AddField("Flex-Rank", currentSummoner.FlexRank.ToString()).
						SetThumbnail(rankTierURL).
						SetImage("attachment://lastgameimage.png").
						SetColor(color).InlineAllFields().MessageEmbed

					messageSend := &discordgo.MessageSend{
						Embeds: []*discordgo.MessageEmbed{embedmessage},
						Files: []*discordgo.File{
							{
								Name:   "lastgameimage.png",
								Reader: lastgameimage,
							},
						},
					}

					_, err = discordSession.ChannelMessageSendComplex(channelID, messageSend)
					if err != nil {
						log.Printf("Failed to send embed message to Discord channel: %v", err)
						return err
					}
				}
			}
		}

		// Update the stored rank
		switch rankType {
		case "Solo":
			summoner.SoloRank = currentSummoner.SoloRank
		case "Flex":
			summoner.FlexRank = currentSummoner.FlexRank
		}
		summoner.Updated = time.Now()
		return nil
	} else {
		return nil
	}
}

// hasSummonerChanged compares two summoners and returns true if they are different
func hasSummonerChanged(oldSummoner, newSummoner *summoner.Summoner) bool {
	return oldSummoner.Name != newSummoner.Name ||
		oldSummoner.TagLine != newSummoner.TagLine ||
		oldSummoner.AccountID != newSummoner.AccountID ||
		oldSummoner.ID != newSummoner.ID ||
		oldSummoner.PUUID != newSummoner.PUUID ||
		oldSummoner.ProfileIconID != newSummoner.ProfileIconID ||
		oldSummoner.SoloRank != newSummoner.SoloRank ||
		oldSummoner.FlexRank != newSummoner.FlexRank
}

// CheckForOngoingGames checks for ongoing games for all registered summoners and sends a message to the Discord channel if a new ongoing game is detected.
func CheckForOngoingGames(discordSession *discordgo.Session, channelID string, summoners map[string]*summoner.Summoner) {
	log.Printf("Checking for ongoing games")

	// Iterate over each summoner to check for ongoing matches
	for _, summoner := range summoners {
		// Load ongoing matches from file
		ongoingMatches, err := databaseHelper.LoadOngoingMatchFromFile()
		if err != nil {
			log.Printf("Failed to load ongoing matches: %v", err)
			return
		}
		// Fetch ongoing match data for the summoner
		currentOngoingMatch, err := apiHelper.GetOngoingMatchByPUUID(summoner.PUUID, os.Getenv("ROPT_API_TOKEN"))
		if err != nil {
			log.Printf("Failed to fetch ongoing match data: %v", err)
			continue
		}

		// If there is no ongoing match, log and continue
		if currentOngoingMatch == nil {
			continue
		}

		if currentOngoingMatch.GameType == "UNRANKED" {
			continue
		}

		// Check if the match is already stored
		if len(ongoingMatches) != 0 {
			if ongoingMatches[currentOngoingMatch.GameID] != nil {
				continue
			} else {
				// Save the new ongoing match
				err = databaseHelper.SaveOngoingMatchToFile(currentOngoingMatch)
				if err != nil {
					log.Printf("Failed to save ongoing match: %v", err)
					continue
				}
			}
		}

		// Check if the current ongoing match is already known
		matchKnown := false
		for _, match := range ongoingMatches {
			if currentOngoingMatch.GameID == match.GameID {
				matchKnown = true
				continue
			}
		}

		// If the match is not known, send a message to the Discord channel
		if !matchKnown {
			// Iterate over each team in the ongoing match
			for teamid, team := range currentOngoingMatch.Teams {
				// Iterate over each participant in the team
				for _, participant := range team.Participants {
					// Iterate over each summoner to see if the participant is known as summoner
					for _, s := range summoners {
						if participant.Summoner.PUUID == s.PUUID {
							var rank rank.Rank
							if currentOngoingMatch.GameType == "Solo/Duo" {
								rank = s.SoloRank
							} else if currentOngoingMatch.GameType == "Flex" {
								rank = s.FlexRank
							} else if currentOngoingMatch.GameType == "UNRANKED" {
								continue
							}

							enemyteamid := 1
							if teamid == 1 {
								enemyteamid = 0
							}

							rankTier := strings.Split(rank.ToString(), " ")[0]
							rankTier = strings.ToLower(rankTier)

							// Get the ranked picture URL
							rankTierURL := cdragon.GetRankedPictureURL(rankTier)

							// Send a message to the Discord channel
							embedmessage := embed.NewEmbed().
								SetAuthor(rank.ToString(), rankTierURL, fmt.Sprintf("https://www.op.gg/summoners/euw/%v-%v", s.Name, s.TagLine)).
								SetTitle(fmt.Sprintf("A %v-Match has started!", currentOngoingMatch.GameType)).
								AddField("Your Team Average Rank", currentOngoingMatch.Teams[teamid].AverageRank().ToString()).
								AddField("Enemy Team Average Rank", currentOngoingMatch.Teams[enemyteamid].AverageRank().ToString()).
								SetThumbnail(cdragon.GetChampionSquareURL(participant.ChampionID)).
								SetFooter(s.GetNameTag(), cdragon.GetProfileIconURL(s.ProfileIconID), fmt.Sprintf("https://www.op.gg/summoners/euw/%v-%v", s.Name, s.TagLine), fmt.Sprintf("https://www.op.gg/summoners/euw/%v-%v", s.Name, s.TagLine)).
								InlineAllFields().MessageEmbed

							messageSend := &discordgo.MessageSend{
								Embeds: []*discordgo.MessageEmbed{embedmessage},
							}

							_, err := discordSession.ChannelMessageSendComplex(channelID, messageSend)
							if err != nil {
								log.Printf("Failed to send embed message to Discord channel: %v", err)
							}
						}
					}
				}
			}
		}
	}
}

// CheckForUpdates continuously checks for rank updates for all registered summoners
func CheckForUpdates() {
	for {
		// Load summoners from file
		summoners, err := databaseHelper.LoadSummonersFromFile()
		if err != nil {
			log.Printf("Failed to load summoners: %v", err)
			time.Sleep(2 * time.Minute)
			continue
		}

		// Flag to track if any summoner has been updated
		updated := false

		CheckForOngoingGames(discordSession, channelID, summoners)

		for name, summoner := range summoners {
			currentSummoner, err := apiHelper.GetSummonerByPUUID(summoner.PUUID)
			if err != nil {
				log.Printf("Failed to fetch summoner data for %v: %v", name, err)
				continue
			}

			log.Printf("Checking for rank updates for summoner: %v", name)

			// Compare summoners and process only if something changed
			if hasSummonerChanged(summoner, currentSummoner) {
				updated = true
				var err error
				err = checkAndSendRankUpdate(discordSession, channelID, currentSummoner, summoner, "Solo")
				if err != nil {
					log.Printf("Failed to check and send rank update for Solo: %v", err)
				}
				err = checkAndSendRankUpdate(discordSession, channelID, currentSummoner, summoner, "Flex")
				if err != nil {
					log.Printf("Failed to check and send rank update for Flex: %v", err)
				}
			}
			// Update the summoner data
			summoners[name] = currentSummoner
		}

		// Save the updated summoners to file only if there were updates
		if updated {
			err = databaseHelper.SaveSummonersToFile(summoners)
			if err != nil {
				log.Printf("Failed to save summoners: %v", err)
			}
		}
	}
}
