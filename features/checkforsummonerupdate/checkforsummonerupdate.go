package checkforsummonerupdate

import (
	"fmt"
	"log"
	"time"

	apiHelper "discord-bot/api-helper"
	databaseHelper "discord-bot/database-helper"

	"github.com/bwmarrin/discordgo"
)

var discordSession *discordgo.Session
var channelID string

// Initialize sets the Discord session and channel ID for sending messages
func Initialize(session *discordgo.Session, chID string) {
	discordSession = session
	channelID = chID
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

		for name, summoner := range summoners {
			currentSummoner, err := apiHelper.GetSummonerByPUUID(summoner.PUUID, summoner.Name, summoner.TagLine)
			if err != nil {
				log.Printf("Failed to fetch summoner data for %v: %v", name, err)
				continue
			}

			// Check for SoloRank updates
			if currentSummoner.SoloRank != summoner.SoloRank {
				rankChange := currentSummoner.SoloRank - summoner.SoloRank
				rankChangeString := ""
				if rankChange < 0 {
					rankChangeString = fmt.Sprintf("%v", rankChange)
				} else {
					rankChangeString = fmt.Sprintf("+%v", rankChange)
				}
				message := fmt.Sprintf("Summoner %v has a new SoloRank: %v. (%v LP)", summoner.GetNameTag(), currentSummoner.SoloRank.ToString(), rankChangeString)
				log.Println(message)
				if discordSession != nil && channelID != "" {
					_, err := discordSession.ChannelMessageSend(channelID, message)
					if err != nil {
						log.Printf("Failed to send message to Discord channel: %v", err)
					}
				}

				// Update the stored SoloRank
				summoner.LastSoloRank = summoner.SoloRank
				summoner.SoloRank = currentSummoner.SoloRank
				summoner.Updated = time.Now()
			}

			// Check for FlexRank updates
			if currentSummoner.FlexRank != summoner.FlexRank {
				rankChange := currentSummoner.FlexRank - summoner.FlexRank
				rankChangeString := ""
				if rankChange < 0 {
					rankChangeString = fmt.Sprintf("%v", rankChange)
				} else {
					rankChangeString = fmt.Sprintf("+%v", rankChange)
				}
				message := fmt.Sprintf("Summoner %v has a new FlexRank: %v. (%v LP)", summoner.GetNameTag(), currentSummoner.FlexRank.ToString(), rankChangeString)
				log.Println(message)
				if discordSession != nil && channelID != "" {
					_, err := discordSession.ChannelMessageSend(channelID, message)
					if err != nil {
						log.Printf("Failed to send message to Discord channel: %v", err)
					}
				}

				// Update the stored FlexRank
				summoner.LastFlexRank = summoner.FlexRank
				summoner.FlexRank = currentSummoner.FlexRank
				summoner.Updated = time.Now()
			}
		}

		// Save the updated summoners to file
		err = databaseHelper.SaveSummonersToFile(summoners)
		if err != nil {
			log.Printf("Failed to save summoners: %v", err)
		}

		// Sleep for a specified interval before checking again
		time.Sleep(2 * time.Minute)
	}
}
