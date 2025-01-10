package offboarding

import (
	databaseHelper "discord-bot/internal/app/helper/database"
	"errors"
	"strings"
)

// DeleteSummoner deletes a summoner by name.
func DeleteSummoner(summonerNameTag string) error {
	summoners, err := databaseHelper.LoadSummonersFromFile()
	if err != nil {
		return err
	}

	// Check if the summoner exists in the map
	var foundKey string
	for key := range summoners {
		if strings.EqualFold(key, summonerNameTag) {
			foundKey = key
			break
		}
	}

	if foundKey == "" {
		return errors.New("summoner not found")
	}

	if foundKey != "" {
		delete(summoners, foundKey)
	}

	// Save the updated map back to the file
	err = databaseHelper.SaveSummonersToFile(summoners)
	if err != nil {
		return err
	}

	return nil
}
