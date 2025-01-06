package checkforsummonerupdate

import (
    "fmt"
    "log"
    "time"

    "discord-bot/api-helper"
    "discord-bot/database-helper"

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

            if currentSummoner.Rank != summoner.Rank {
                // Post a message about the rank change
                message := fmt.Sprintf("Summoner %v has a new rank: %v (was %v)", summoner.Name, currentSummoner.Rank, summoner.Rank)
                log.Println(message)
                if discordSession != nil && channelID != "" {
                    _, err := discordSession.ChannelMessageSend(channelID, message)
                    if err != nil {
                        log.Printf("Failed to send message to Discord channel: %v", err)
                    }
                }

                // Update the stored rank
                summoner.Rank = currentSummoner.Rank
                summoner.LastRank = currentSummoner.Rank
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