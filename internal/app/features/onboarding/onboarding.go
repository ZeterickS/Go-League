package onboarding

import (
	apiHelper "discord-bot/internal/app/helper/api"
	databaseHelper "discord-bot/internal/app/helper/database"
	"discord-bot/types/embed"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// OnboardSummoner fetches summoner data by tag and saves it to the database
func OnboardSummoner(name, tagLine string) (*discordgo.MessageSend, error) {

	summoners, err := databaseHelper.LoadSummonersFromFile()
	if err != nil {
		return nil, fmt.Errorf("failed to load summoner data: %v", err)
	}

	// Convert name and tagLine to lowercase for case-insensitive comparison
	nameTag := fmt.Sprintf("%s#%s", name, tagLine)
	lowerNameTag := strings.ToLower(nameTag)

	// Check if the summoner already exists in a case-insensitive manner
	for existingNameTag := range summoners {
		if strings.ToLower(existingNameTag) == lowerNameTag {
			return nil, fmt.Errorf("summoner with name %s already exists", existingNameTag)
		}
	}

	summonerData, err := apiHelper.GetSummonerByTag(name, tagLine)
	if err != nil {
		if summonerData.SoloRank == 0 && summonerData.FlexRank == 0 {
			log.Printf("No rank available for summoner: %v", nameTag)
		} else {
			return nil, fmt.Errorf("failed to fetch summoner data: %v", err)
		}
	}

	summoners[summonerData.GetNameTag()] = summonerData

	// Better output of the summoners list
	for name, summoner := range summoners {
		log.Printf("Summoner: %s, Data: %+v\n", name, summoner)
	}

	err = databaseHelper.SaveSummonersToFile(summoners)
	if err != nil {
		return nil, fmt.Errorf("failed to save summoner data: %v", err)
	}

	// Prepare the embed message with profile icon as thumbnail
	profileIconPath := fmt.Sprintf("assets/profileicon/%v.png", summonerData.ProfileIconID)
	profileFile, err := os.Open(profileIconPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open profile icon file: %v", err)
	}
	defer profileFile.Close()

	embedMessage := embed.NewEmbed().
		SetAuthor(summonerData.Name, fmt.Sprintf("attachment://%v.png", summonerData.ProfileIconID), "").
		SetTitle("Summoner Onboarded").
		SetDescription(fmt.Sprintf("Summoner %v is now registered", summonerData.GetNameTag())).
		SetThumbnail(fmt.Sprintf("attachment://%v.png", summonerData.ProfileIconID)).
		SetColor(0x00ff00).MessageEmbed

	messageSend := &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{embedMessage},
		Files: []*discordgo.File{
			{
				Name:        fmt.Sprintf("%v.png", summonerData.ProfileIconID),
				ContentType: "image/png",
				Reader:      profileFile,
			},
		},
	}

	return messageSend, nil
}
