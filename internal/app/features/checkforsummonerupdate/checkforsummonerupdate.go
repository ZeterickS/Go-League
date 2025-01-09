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
	switch rankType {
	case "Solo":
		currentRank = currentSummoner.SoloRank
		previousRank = summoner.SoloRank
	case "Flex":
		currentRank = currentSummoner.FlexRank
		previousRank = summoner.FlexRank
	default:
		log.Printf("Unknown rank type: %v", rankType)
		return fmt.Errorf("unknown rank type: %v", rankType)
	}

	if currentRank != previousRank {
		rankChange := currentRank - previousRank
		rankChangeString := ""
		message := ""
		if rankChange < 0 {
			rankChangeString = fmt.Sprintf("%v", rankChange)
			message = fmt.Sprintf("%v has lost a Game: %v. (%v LP)", summoner.GetNameTag(), currentRank.ToString(), rankChangeString)
		} else {
			rankChangeString = fmt.Sprintf("+%v", rankChange)
			message = fmt.Sprintf("%v has won a Game: %v. (%v LP)", summoner.GetNameTag(), currentRank.ToString(), rankChangeString)
		}
		color := 0x00ff00 // Green color for LP gain
		if rankChange < 0 {
			color = 0xff0000 // Red color for LP loss
		}

		rankTier := strings.Split(currentRank.ToString(), " ")[0]
		rankTier = strings.ToLower(rankTier)
		fmt.Println(rankTier)

		imagePath := fmt.Sprintf("assets/rank_images/%v.png", rankTier)
		rankfile, err := os.Open(imagePath)
		if err != nil {
			log.Printf("Failed to open image file: %v", err)
			return err
		}
		defer rankfile.Close()

		embedmessage := embed.NewEmbed().
			SetAuthor(currentSummoner.Name, cdragon.GetProfileIconURL(currentSummoner.ProfileIconID)).
			SetTitle(fmt.Sprintf("Rank Update %v", rankType)).
			SetDescription(message).
			AddField("Rank Change", rankChangeString).
			SetThumbnail(fmt.Sprintf("attachment://%v.png", rankTier)).
			SetColor(color).MessageEmbed

		messageSend := &discordgo.MessageSend{
			Embeds: []*discordgo.MessageEmbed{embedmessage},
			Files: []*discordgo.File{
				{
					Name:        fmt.Sprintf("%v.png", rankTier),
					ContentType: "image/png",
					Reader:      rankfile,
				},
			},
		}

		_, err = discordSession.ChannelMessageSendComplex(channelID, messageSend)
		if err != nil {
			log.Printf("Failed to send embed message with file to Discord channel: %v", err)
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
			currentSummoner, err := apiHelper.GetSummonerByPUUID(summoner.PUUID, summoner.Name, summoner.TagLine)
			log.Printf("Current summoner data for %v: %+v", name, currentSummoner)
			if err != nil {
				log.Printf("Failed to fetch summoner data for %v: %v", name, err)
				continue
			}

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
					log.Printf("Failed to check and send rank update for Solo: %v", err)
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
