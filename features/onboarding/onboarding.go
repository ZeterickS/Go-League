package onboarding

import (
	apiHelper "discord-bot/api-helper"
	databaseHelper "discord-bot/database-helper"
	"discord-bot/types/summoner"
	"fmt"
	"log"
)

// OnboardSummoner fetches summoner data by tag and saves it to the database
func OnboardSummoner(name, tagLine string) (*summoner.Summoner, error) {

	summoners, err := databaseHelper.LoadSummonersFromFile()
	if err != nil {
		return nil, fmt.Errorf("failed to load summoner data: %v", err)
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
