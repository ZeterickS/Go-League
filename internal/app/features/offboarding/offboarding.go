package offboarding

import (
	databaseHelper "discord-bot/internal/app/helper/database"
	"discord-bot/internal/logger"

	"go.uber.org/zap"
)

// DeleteSummoner deletes a summoner by name.
func DeleteSummoner(summonerNameTag, channelID string) error {
	err := databaseHelper.DeleteChannelForSummoner(summonerNameTag, channelID)
	if err != nil {
		logger.Logger.Error("Failed to delete channel for summoner", zap.String("summonerNameTag", summonerNameTag), zap.String("channelID", channelID), zap.Error(err))
		return err
	}

	logger.Logger.Info("Successfully deleted channel for summoner", zap.String("summonerNameTag", summonerNameTag), zap.String("channelID", channelID))
	return nil
}

func DeleteChannel(channelID string) error {
	err := databaseHelper.DeleteChannel(channelID)
	if err != nil {
		logger.Logger.Error("Failed to delete channel", zap.String("channelID", channelID), zap.Error(err))
		return err
	}

	logger.Logger.Info("Successfully deleted channel", zap.String("channelID", channelID))
	return nil
}

func DeleteGuild(guildID string) error {
	err := databaseHelper.DeleteGuild(guildID)
	if err != nil {
		logger.Logger.Error("Failed to delete guild", zap.String("guildID", guildID), zap.Error(err))
		return err
	}

	logger.Logger.Info("Successfully deleted guild", zap.String("guildID", guildID))
	return nil
}
