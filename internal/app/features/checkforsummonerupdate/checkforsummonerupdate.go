package checkforsummonerupdate

import (
	"fmt"
	"log"
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

// Initialize sets the Discord session and channel ID for sending messages
func Initialize(session *discordgo.Session) {
	discordSession = session
}

func checkAndSendRankUpdate(summoner summoner.Summoner) error {
	var pretttyRank, rankType string = "", ""

	newSoloRank, newFlexRank, err := apiHelper.GetSummonerRank(summoner.ID, summoner.Region)
	if err != nil {
		log.Printf("Failed to fetch summoner by PUUID: %v", err)
		return err
	}

	if newSoloRank != summoner.SoloRank {
		pretttyRank = "Solo/Duo"
		rankType = "Solo"
	} else if newFlexRank != summoner.FlexRank {
		pretttyRank = "Flex"
		rankType = "Flex"
	} else {
		return nil
	}

	lastmatchid, err := apiHelper.GetLastRankedMatchIDbyPUUID(summoner.PUUID)
	log.Printf("Last ranked match ID: %v", lastmatchid)

	if err != nil {
		log.Printf("Failed to fetch last ranked match ID: %v", err)
		return err
	}

	matchExists, err := databaseHelper.IsMatchExists(lastmatchid)
	if err != nil {
		log.Printf("Failed to check if match exists: %v", err)
	}

	if matchExists {
		return nil
	}

	lastMatch, err := apiHelper.GetMatchByID(lastmatchid)
	log.Printf("Last match: %v", lastMatch)

	if err != nil {
		log.Printf("Failed to fetch last match: %v", err)
		return err
	}

	if lastMatch == nil {
		log.Printf("Last match is nil")
		return fmt.Errorf("last match is nil")
	}

	for i := 0; i < len(lastMatch.Teams); i++ {
		for _, participant := range lastMatch.Teams[i].Participants {
			// If Participant is not mapped to any channel, skip
			participantIsMapped, err := databaseHelper.IsSummonerMappedToAnyChannel(participant.Summoner.PUUID)
			if !participantIsMapped || err != nil {
				continue
			}

			// Dont call API if Summoner has already got the Rank Update Call
			var newparticipantSoloRank, newparticipantFlexRank rank.Rank

			if summoner.PUUID == participant.Summoner.PUUID {
				newparticipantSoloRank = newSoloRank
				newparticipantFlexRank = newFlexRank
			} else {
				newparticipantSoloRank, newparticipantFlexRank, err = apiHelper.GetSummonerRank(participant.Summoner.ID, participant.Summoner.Region)
				if err != nil {
					log.Printf("Failed to fetch summoner by PUUID: %v", err)
					continue
				}
			}

			var currentRank, rankChange rank.Rank
			rankChangeString := ""

			switch rankType {
			case "Solo":
				currentRank = newparticipantSoloRank
				rankChange = newparticipantSoloRank - participant.Summoner.SoloRank
			case "Flex":
				currentRank = newparticipantFlexRank
				rankChange = newparticipantFlexRank - participant.Summoner.FlexRank
			}

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

			knownChannels, err := databaseHelper.GetChannelsForSummoner(participant.Summoner.PUUID)
			if err != nil {
				log.Printf("Failed to get channel by summoner PUUID: %v", err)
				continue
			}
			lastgameimage, err := gametoimage.GameToImage(participant)
			if err != nil {
				log.Printf("Failed to generate game image: %v", err)
				continue
			}
			for _, knownChannel := range knownChannels {
				embedmessage := embed.NewEmbed().
					SetAuthor(participant.Summoner.GetNameTag(), cdragon.GetProfileIconURL(participant.Summoner.ProfileIconID), fmt.Sprintf("https://www.op.gg/summoners/euw/%v-%v", participant.Summoner.Name, participant.Summoner.TagLine), fmt.Sprintf("https://www.op.gg/summoners/euw/%v-%v", participant.Summoner.Name, participant.Summoner.TagLine)).
					SetTitle(fmt.Sprintf("%v-Rank Update | %v LP", pretttyRank, rankChangeString)).
					AddField("Solo/Duo-Rank", participant.Summoner.SoloRank.ToString()).
					AddField("Flex-Rank", participant.Summoner.FlexRank.ToString()).
					SetThumbnail(cdragon.GetChampionSquareURL(participant.ChampionID)).
					SetImage("attachment://lastgameimage.png").
					SetFooter(currentRank.ToString(), rankTierURL, fmt.Sprintf("https://www.op.gg/summoners/euw/%v-%v", participant.Summoner.Name, participant.Summoner.TagLine)).
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

				_, err = discordSession.ChannelMessageSendComplex(knownChannel, messageSend)
				if err != nil {
					log.Printf("Failed to send embed message to Discord channel: %v", err)
					lastgameimage.Close()
					return err
				}

				lastgameimage.Close()
			}
			participant.Summoner.SoloRank = newparticipantSoloRank
			participant.Summoner.FlexRank = newparticipantFlexRank
			participant.Summoner.Updated = time.Now()
			err = databaseHelper.SaveSummonerToDB(participant.Summoner)
			if err != nil {
				log.Printf("Failed to save summoner to DB: %v", err)
				return err
			}
		}
	}
	// Save the match to the database
	oldGameID := strings.Split(lastmatchid, "_")[1]
	err = databaseHelper.UpdateOngoingToFinishedGame(oldGameID, lastMatch)
	if err != nil {
		log.Printf("Failed to update ongoing game to finished: %v", err)
		return err
	}
	return nil
}

// CheckForOngoingGames checks for ongoing games for all registered summoners and sends a message to the Discord channel if a new ongoing game is detected.
func checkForOngoingGames(checksummoner *summoner.Summoner) {
	log.Printf("Checking for ongoing games for summoner: %v", checksummoner.GetNameTag())
	log.Printf("Summoner PUUID: %v", checksummoner.PUUID)

	ongoingMatch, err := apiHelper.GetOngoingMatchByPUUID(checksummoner.PUUID, checksummoner.Region)
	if err != nil {
		log.Printf("Failed to get ongoing match by PUUID: %v", err)
		return
	}

	if ongoingMatch != nil {
		log.Printf("Ongoing match detected for summoner: %v", checksummoner.GetNameTag())
		// Save the new ongoing match
		err = databaseHelper.SaveOngoingMatchToDB(ongoingMatch)
		if err != nil {
			log.Printf("Failed to save ongoing match: %v", err)
			return
		}
	} else {
		log.Printf("No ongoing match detected for summoner: %v", checksummoner.GetNameTag())
		return
	}

	if ongoingMatch.GameType == "UNRANKED" {
		log.Printf("Ongoing match is unranked for summoner: %v", checksummoner.GetNameTag())
		return
	}

	// Iterate over each team in the ongoing match
	for teamid, team := range ongoingMatch.Teams {
		// Iterate over each participant in the team
		for _, participant := range team.Participants {
			// Check if the summoner is registered
			summonerMapped, err := databaseHelper.IsSummonerMappedToAnyChannel(participant.Summoner.PUUID)
			if err != nil {
				log.Printf("Failed to check if summoner is registered: %v", err)
				continue
			}
			if summonerMapped {
				log.Printf("Summoner %v is mapped to a channel", participant.Summoner.GetNameTag())

				var rank rank.Rank
				if ongoingMatch.GameType == "Solo/Duo" {
					rank = participant.Summoner.SoloRank
				} else if ongoingMatch.GameType == "Flex" {
					rank = participant.Summoner.FlexRank
				} else if ongoingMatch.GameType == "UNRANKED" {
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

				knownChannels, err := databaseHelper.GetChannelsForSummoner(participant.Summoner.PUUID)
				if err != nil {
					log.Printf("Failed to get channel by summoner PUUID: %v", err)
					continue
				}

				for _, knownChannel := range knownChannels {
					log.Printf("Sending ongoing match notification to channel: %v", knownChannel)
					// Send a message to the Discord channel
					embedmessage := embed.NewEmbed().
						SetAuthor(participant.Summoner.GetNameTag(), cdragon.GetProfileIconURL(participant.Summoner.ProfileIconID), fmt.Sprintf("https://www.op.gg/summoners/euw/%v-%v", participant.Summoner.Name, participant.Summoner.TagLine), fmt.Sprintf("https://www.op.gg/summoners/euw/%v-%v", participant.Summoner.Name, participant.Summoner.TagLine)).
						SetTitle(fmt.Sprintf("A %v-Match has started!", ongoingMatch.GameType)).
						AddField("Your Team Average Rank", ongoingMatch.Teams[teamid].AverageRank().ToString()).
						AddField("Enemy Team Average Rank", ongoingMatch.Teams[enemyteamid].AverageRank().ToString()).
						SetThumbnail(cdragon.GetChampionSquareURL(participant.ChampionID)).
						SetFooter(rank.ToString(), rankTierURL, fmt.Sprintf("https://www.op.gg/summoners/euw/%v-%v", participant.Summoner.Name, participant.Summoner.TagLine)).
						InlineAllFields().MessageEmbed

					messageSend := &discordgo.MessageSend{
						Embeds: []*discordgo.MessageEmbed{embedmessage},
					}

					_, err := discordSession.ChannelMessageSendComplex(knownChannel, messageSend)
					if err != nil {
						log.Printf("Failed to send embed message to Discord channel: %v", err)
					}
				}
			} else {
				log.Printf("Summoner %v is not mapped to any channel", participant.Summoner.GetNameTag())
			}
		}
	}
}

// CheckForUpdates continuously checks for rank updates for all registered summoners
func CheckForUpdates() {
	for {
		//load oldest summoner from database
		summonerPUUID, err := databaseHelper.GetOldestSummonerWithChannel()
		if err != nil {
			//log.Printf("Failed to get oldest summoner: %v", err)
			continue
		}

		log.Printf("Checking for updates for summoner: %v", summonerPUUID)

		oldestsummoner, err := databaseHelper.GetSummonerByPUUIDFromDB(summonerPUUID)
		if err != nil {
			log.Printf("Failed to get summoner by PUUID: %v", err)
			continue
		}

		log.Printf("Checking for updates for summoner: %v", oldestsummoner.GetNameTag())
		log.Printf("Oldest Summoner is: %v", oldestsummoner.Updated)
		log.Printf("Time since last update: %v", time.Since(oldestsummoner.Updated))
		log.Printf("Summoner details: %+v", oldestsummoner)

		// Compare summoners and process only if something changed
		checkAndSendRankUpdate(*oldestsummoner)

		checkForOngoingGames(oldestsummoner)

		databaseHelper.UpdateSummonerTimestamp(oldestsummoner.PUUID)
	}
}
