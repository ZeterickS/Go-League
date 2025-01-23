package onboarding

import (
	apiHelper "discord-bot/internal/app/helper/api"
	"discord-bot/internal/app/helper/cdragon"
	databaseHelper "discord-bot/internal/app/helper/database"
	"discord-bot/internal/logger"
	"discord-bot/types/embed"
	"discord-bot/types/summoner"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

// OnboardSummoner fetches summoner data by tag and saves it to the database
func OnboardSummoner(name, tagLine, region, channelID string) (*discordgo.MessageEmbed, error) {
	logger.Logger.Info("Onboarding summoner", zap.String("name", name), zap.String("tagLine", tagLine), zap.String("region", region), zap.String("channelID", channelID))
	var summoner *summoner.Summoner

	// Sanity check for name and tagLine to ensure they are URL-safe
	if strings.ContainsAny(name, "!@#$%^&*()+=[]{}|\\;:'\",<>/?") || strings.ContainsAny(tagLine, " !@#$%^&*()+=[]{}|\\;:'\",<>/?") {
		logger.Logger.Warn("Invalid characters in name or tagLine", zap.String("name", name), zap.String("tagLine", tagLine))
		return nil, fmt.Errorf("name or tagLine contains invalid characters")
	}

	// Sanity check for SQL injection
	if strings.ContainsAny(name, "'\";--") || strings.ContainsAny(tagLine, "'\";--") {
		logger.Logger.Warn("SQL injection characters in name or tagLine", zap.String("name", name), zap.String("tagLine", tagLine))
		return nil, fmt.Errorf("name or tagLine contains SQL injection characters")
	}

	summonerExists, err := databaseHelper.SummonerExists(name, tagLine, region)
	if err != nil {
		logger.Logger.Error("Failed to check if summoner exists", zap.Error(err))
		return nil, fmt.Errorf("failed to check if summoner exists: %v", err)
	}

	if summonerExists {
		logger.Logger.Info("Summoner already exists", zap.String("name", name), zap.String("tagLine", tagLine), zap.String("region", region))
		summoner, err = databaseHelper.GetDBSummonerByName(name, tagLine, region)
		if err != nil {
			logger.Logger.Error("Failed to fetch summoner data", zap.Error(err))
			return nil, fmt.Errorf("failed to fetch summoner data: %v", err)
		}
	} else {
		summoner, err = apiHelper.GetSummonerByTag(name, tagLine, region)
		if err != nil {
			logger.Logger.Error("Failed to fetch summoner data", zap.Error(err))
			return nil, fmt.Errorf("failed to fetch summoner data: %v", err)
		}
		logger.Logger.Info("Fetched summoner data", zap.Any("summoner", summoner))

		err = databaseHelper.SaveSummonerToDB(*summoner)
		if err != nil {
			return nil, fmt.Errorf("failed to save summoner to database: %v", err)
		}
	}

	err = databaseHelper.SaveChannelForSummoner(summoner.PUUID, channelID)
	if err != nil {
		return nil, fmt.Errorf("summoner already exists in this channel: name=%s, tagLine=%s, region=%s, channelID=%s", name, tagLine, region, channelID)
	}

	embedMessage := embed.NewEmbed().
		SetTitle("Summoner Onboarded").
		SetDescription(fmt.Sprintf("Summoner %v is now registered", summoner.GetNameTag())).
		AddField("Solo-Rank", summoner.SoloRank.ToString()).
		AddField("Flex-Rank", summoner.FlexRank.ToString()).
		SetThumbnail(cdragon.GetProfileIconURL(summoner.ProfileIconID)).
		InlineAllFields().MessageEmbed

	return embedMessage, nil
}
