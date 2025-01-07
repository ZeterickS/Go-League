package checkforsummonerupdate

import (
	"fmt"
	"log"
	"strings"
	"time"

	apiHelper "discord-bot/internal/app/helper/api"
	databaseHelper "discord-bot/internal/app/helper/database"

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
				message := ""
				if rankChange < 0 {
					rankChangeString = fmt.Sprintf("%v", rankChange)
					message = fmt.Sprintf("%v has lost a Game: %v. (%v LP)", summoner.GetNameTag(), currentSummoner.SoloRank.ToString(), rankChangeString)
				} else {
					rankChangeString = fmt.Sprintf("+%v", rankChange)
					message = fmt.Sprintf("%v has lost a Game: %v. (%v LP)", summoner.GetNameTag(), currentSummoner.SoloRank.ToString(), rankChangeString)
				}
				color := 0x00ff00 // Green color for LP gain
				if rankChange < 0 {
					color = 0xff0000 // Red color for LP loss
				}
				rankTier := strings.Split(currentSummoner.SoloRank.ToString(), " ")[0]
				rankTier = strings.ToLower(rankTier)
				fmt.Println(rankTier)
				embed := &discordgo.MessageEmbed{
					Title:       "Rank Update Solo/Duo",
					Description: message,
					Color:       color,
					Thumbnail: &discordgo.MessageEmbedThumbnail{
						URL: fmt.Sprintf("https://raw.communitydragon.org/latest/plugins/rcp-fe-lol-static-assets/global/default/images/ranked-mini-crests/%v.png", rankTier),
					},
				}

				log.Println(embed.Description)
				if discordSession != nil && channelID != "" {
					_, err := discordSession.ChannelMessageSendEmbed(channelID, embed)
					if err != nil {
						log.Printf("Failed to send embed message to Discord channel: %v", err)
					}
				}

				// Update the stored SoloRank
				summoner.LastSoloRank = summoner.SoloRank
				summoner.SoloRank = currentSummoner.SoloRank
				summoner.Updated = time.Now()

				// Check for FlexRank updates
				if currentSummoner.FlexRank != summoner.FlexRank {
					rankChange := currentSummoner.FlexRank - summoner.FlexRank
					rankChangeString := ""
					message := ""
					if rankChange < 0 {
						rankChangeString = fmt.Sprintf("%v", rankChange)
						message = fmt.Sprintf("%v has lost a Game: %v. (%v LP)", summoner.GetNameTag(), currentSummoner.FlexRank.ToString(), rankChangeString)
					} else {
						rankChangeString = fmt.Sprintf("+%v", rankChange)
						message = fmt.Sprintf("%v has lost a Game: %v. (%v LP)", summoner.GetNameTag(), currentSummoner.FlexRank.ToString(), rankChangeString)
					}
					color := 0x00ff00 // Green color for LP gain
					if rankChange < 0 {
						color = 0xff0000 // Red color for LP loss
					}
					rankTier := strings.Split(currentSummoner.FlexRank.ToString(), " ")[0]
					rankTier = strings.ToLower(rankTier)
					fmt.Println(rankTier)
					embed := &discordgo.MessageEmbed{
						Title:       "Rank Update Flex",
						Description: message,
						Color:       color,
						Thumbnail: &discordgo.MessageEmbedThumbnail{
							URL: fmt.Sprintf("https://raw.communitydragon.org/latest/plugins/rcp-fe-lol-static-assets/global/default/images/ranked-mini-crests/%v.png", rankTier),
						},
					}

					log.Println(embed.Description)
					if discordSession != nil && channelID != "" {
						_, err := discordSession.ChannelMessageSendEmbed(channelID, embed)
						if err != nil {
							log.Printf("Failed to send embed message to Discord channel: %v", err)
						}
					}

					// Update the stored FlexRank
					summoner.LastFlexRank = summoner.FlexRank
					summoner.FlexRank = currentSummoner.FlexRank
					summoner.Updated = time.Now()
				}

				log.Println(embed.Description)
				if discordSession != nil && channelID != "" {
					_, err := discordSession.ChannelMessageSendEmbed(channelID, embed)
					if err != nil {
						log.Printf("Failed to send embed message to Discord channel: %v", err)
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
