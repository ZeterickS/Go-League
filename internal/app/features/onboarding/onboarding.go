package onboarding

import (
	apiHelper "discord-bot/internal/app/helper/api"
	"discord-bot/internal/app/helper/cdragon"
	databaseHelper "discord-bot/internal/app/helper/database"
	"discord-bot/types/embed"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// OnboardSummoner fetches summoner data by tag and saves it to the database
func OnboardSummoner(name, tagLine string) (*discordgo.MessageEmbed, error) {

	// Sanity check for name and tagLine to ensure they are URL-safe
	if strings.ContainsAny(name, " !@#$%^&*()+=[]{}|\\;:'\",<>/?") || strings.ContainsAny(tagLine, " !@#$%^&*()+=[]{}|\\;:'\",<>/?") {
		return nil, fmt.Errorf("name or tagLine contains invalid characters")
	}

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
		return nil, fmt.Errorf("failed to fetch summoner data: %v", err)

	}

	summoners[summonerData.GetNameTag()] = summonerData

	err = databaseHelper.SaveSummonersToFile(summoners)
	if err != nil {
		return nil, fmt.Errorf("failed to save summoner data: %v", err)
	}

	embedMessage := embed.NewEmbed().
		SetTitle("Summoner Onboarded").
		SetDescription(fmt.Sprintf("Summoner %v is now registered", summonerData.GetNameTag())).
		AddField("Solo-Rank", summonerData.SoloRank.ToString()).
		AddField("Flex-Rank", summonerData.FlexRank.ToString()).
		SetThumbnail(cdragon.GetProfileIconURL(summonerData.ProfileIconID)).
		InlineAllFields().MessageEmbed

	return embedMessage, nil
}
