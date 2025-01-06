// FILE: offboarding/offboarding.go
package offboarding

import (
    "errors"
    "discord-bot/database-helper"
)

// DeleteSummoner deletes a summoner by name.
func DeleteSummoner(summonerName string) error {
    summoners, err := databaseHelper.LoadSummonersFromFile()
    if err != nil {
        return err
    }

    // Check if the summoner exists in the map
    if _, exists := summoners[summonerName]; !exists {
        return errors.New("summoner not found")
    }

    // Delete the summoner from the map
    delete(summoners, summonerName)

    // Save the updated map back to the file
    err = databaseHelper.SaveSummonersToFile(summoners)
    if err != nil {
        return err
    }

    return nil
}