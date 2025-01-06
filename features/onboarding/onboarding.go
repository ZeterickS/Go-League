package onboarding

import (
	apiHelper "discord-bot/api-helper"
	databaseHelper "discord-bot/database-helper"
	"discord-bot/types"
	"fmt"
	"log"
)

// OnboardSummoner fetches summoner data by tag and saves it to the database
func OnboardSummoner(name, tagLine string) (*types.Summoner, error) {

	summoners, err := databaseHelper.LoadSummonersFromFile()
	if err != nil {
		return nil, fmt.Errorf("failed to load summoner data: %v", err)
	}

	summoner, err := apiHelper.GetSummonerByTag(name, tagLine)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch summoner data: %v", err)
	}

	if _, exists := summoners[summoner.Name]; exists {
		return nil, fmt.Errorf("summoner with name %s already exists", summoner.Name)
	}

	summoners[summoner.Name] = summoner

	// Better output of the summoners list
	for name, summoner := range summoners {
		log.Printf("Summoner: %s, Data: %+v\n", name, summoner)
	}

	err = databaseHelper.SaveSummonersToFile(summoners)
	if err != nil {
		return nil, fmt.Errorf("failed to save summoner data: %v", err)
	}

	return summoner, nil
}