package onboarding

import (
	apiHelper "discord-bot/internal/app/helper/api"
	databaseHelper "discord-bot/internal/app/helper/database"
	"discord-bot/types/summoner"
	"fmt"
	"log"
	"strings"
)

// OnboardSummoner fetches summoner data by tag and saves it to the database
func OnboardSummoner(name, tagLine string) (*summoner.Summoner, error) {

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

	if _, exists := summoners[summonerData.GetNameTag()]; exists {
		return nil, fmt.Errorf("summoner with name %s already exists", summonerData.GetNameTag())
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

	return summonerData, nil
}
