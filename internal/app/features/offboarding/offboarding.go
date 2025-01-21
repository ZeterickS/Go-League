package offboarding

import (
	databaseHelper "discord-bot/internal/app/helper/database"
)

// DeleteSummoner deletes a summoner by name.
func DeleteSummoner(summonerNameTag, channelID string) error {
	err := databaseHelper.DeleteChannelForSummoner(summonerNameTag, channelID)
	if err != nil {
		return err
	}

	return nil
}
