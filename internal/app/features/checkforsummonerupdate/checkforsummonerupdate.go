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

		// Use the URL directly
		embedmessage := embed.NewEmbed().
			SetAuthor(currentSummoner.GetNameTag(), cdragon.GetProfileIconURL(currentSummoner.ProfileIconID), fmt.Sprintf("https://www.op.gg/summoners/euw/%v-%v", currentSummoner.Name, currentSummoner.TagLine)).
			SetTitle(fmt.Sprintf("%v-Rank Update | %v LP", pretttyRank, rankChangeString)).
			AddField("Solo/Duo-Rank", currentSummoner.SoloRank.ToString()).
			AddField("Flex-Rank", currentSummoner.FlexRank.ToString()).
			SetThumbnail(rankTierURL).
			SetColor(color).InlineAllFields().MessageEmbed

		messageSend := &discordgo.MessageSend{
			Embeds: []*discordgo.MessageEmbed{embedmessage},
		}

		_, err := discordSession.ChannelMessageSendComplex(channelID, messageSend)
		if err != nil {
			log.Printf("Failed to send embed message to Discord channel: %v", err)
			return err
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

// CheckForOngoingGame checks for an ongoing game for a specific summoner and sends a message to the Discord channel if an ongoing game, which is not already stored in the database, is detected.
func CheckForOngoingGame(discordSession *discordgo.Session, channelID string, summoner *summoner.Summoner) {
	// Load ongoing matches from file
	ongoingMatches, err := databaseHelper.LoadOngoingFromFile()
	if err != nil {
		log.Printf("Failed to load ongoing matches: %v", err)
		return
	}

	// Fetch ongoing match data for the summoner
	currentOngoingMatch, err := apiHelper.GetOngoingMatchByPUUID(summoner.PUUID, os.Getenv("ROPT_API_TOKEN"))
	if err != nil {
		log.Printf("Failed to fetch ongoing match data: %v", err)
		return
	}

	// If there is no ongoing match, log and return
	if currentOngoingMatch == nil {
		log.Printf("No ongoing game found for summoner: %s", summoner.Name)
		return
	}

	// Check if the match is already stored
	if len(ongoingMatches) != 0 {
		if ongoingMatches[currentOngoingMatch.GameID] != nil {
			return
		} else {
			// Save the new ongoing match
			ongoingMatches[currentOngoingMatch.GameID] = currentOngoingMatch
			err = databaseHelper.SaveOngoingToFile(currentOngoingMatch)
			if err != nil {
				log.Printf("Failed to save ongoing match: %v", err)
				return
			}
			log.Printf("New ongoing game saved for summoner: %s", summoner.Name)
		}
	}

	var championID int
	for _, participant := range currentOngoingMatch.Teams[0].Participants {
		if participant.Summoner.PUUID == summoner.PUUID {
			championID = participant.ChampionID
			break
		}
	}

	var rank rank.Rank
	if currentOngoingMatch.GameType == "Solo/Duo" {
		rank = summoner.SoloRank
	} else if currentOngoingMatch.GameType == "Flex" {
		rank = summoner.FlexRank
	}

	rankTier := strings.Split(rank.ToString(), " ")[0]
	rankTier = strings.ToLower(rankTier)
	fmt.Println(rankTier)

	// Get the ranked picture URL
	rankTierURL := cdragon.GetRankedPictureURL(rankTier)

	// Check if the current ongoing match is already known
	matchKnown := false
	for _, match := range ongoingMatches {
		if currentOngoingMatch.GameID == match.GameID {
			matchKnown = true
			break
		}
	}

	// If there is an ongoing match and it's not already stored in the database
	if !matchKnown {
		// Send a message to the Discord channel
		embedmessage := embed.NewEmbed().
			SetAuthor(summoner.GetNameTag(), cdragon.GetProfileIconURL(summoner.ProfileIconID), fmt.Sprintf("https://www.op.gg/summoners/euw/%v-%v", summoner.Name, summoner.TagLine)).
			SetTitle(fmt.Sprintf("A %v-Rankmatch has started!", currentOngoingMatch.GameType)).
			AddField("Your Team Average Rank", currentOngoingMatch.Teams[0].AverageRank().ToString()).
			AddField("Enemy Team Average Rank", currentOngoingMatch.Teams[1].AverageRank().ToString()).
			SetThumbnail(cdragon.GetChampionSquareURL(championID)).
			SetFooter(rank.ToString(), rankTierURL).
			InlineAllFields().MessageEmbed

		messageSend := &discordgo.MessageSend{
			Embeds: []*discordgo.MessageEmbed{embedmessage},
		}

		_, err := discordSession.ChannelMessageSendComplex(channelID, messageSend)
		if err != nil {
			log.Printf("Failed to send embed message to Discord channel: %v", err)
		}

		// Save the ongoing match to file
		err = databaseHelper.SaveOngoingToFile(currentOngoingMatch)
		if err != nil {
			log.Printf("Failed to save ongoing match to file: %v", err)
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

		for name, summoner := range summoners {
			currentSummoner, err := apiHelper.GetSummonerByPUUID(summoner.PUUID)
			if err != nil {
				log.Printf("Failed to fetch summoner data for %v: %v", name, err)
				continue
			}

			CheckForOngoingGame(discordSession, channelID, currentSummoner)

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

		// Sleep for a specified interval before checking again
		time.Sleep(30 * time.Second)
	}
}
