package offboarding

import (
    "errors"
    "discord-bot/database-helper"
)

// DeleteSummoner deletes a summoner by name.
func DeleteSummoner(summonerNameTag string) error {
    summoners, err := databaseHelper.LoadSummonersFromFile()
    if err != nil {
        return err
    }

    // Check if the summoner exists in the map
    if _, exists := summoners[summonerNameTag]; !exists {
        return errors.New("summoner not found")
    }

    // Delete the summoner from the map
    delete(summoners, summonerNameTag)

    // Save the updated map back to the file
    err = databaseHelper.SaveSummonersToFile(summoners)
    if err != nil {
        return err
    }

    return nil
}