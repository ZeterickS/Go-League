package onboarding

import (
	apiHelper "discord-bot/internal/app/helper/api"
	"discord-bot/internal/app/helper/cdragon"
	databaseHelper "discord-bot/internal/app/helper/database"
	"discord-bot/types/embed"
	"discord-bot/types/summoner"
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// OnboardSummoner fetches summoner data by tag and saves it to the database
func OnboardSummoner(name, tagLine, region, channelID string) (*discordgo.MessageEmbed, error) {
	log.Printf("Onboarding summoner: name=%s, tagLine=%s, region=%s, channelID=%s", name, tagLine, region, channelID)
	var summoner *summoner.Summoner

	// Sanity check for name and tagLine to ensure they are URL-safe
	if strings.ContainsAny(name, "!@#$%^&*()+=[]{}|\\;:'\",<>/?") || strings.ContainsAny(tagLine, " !@#$%^&*()+=[]{}|\\;:'\",<>/?") {
		log.Printf("Invalid characters in name or tagLine: name=%s, tagLine=%s", name, tagLine)
		return nil, fmt.Errorf("name or tagLine contains invalid characters")
	}

	// Sanity check for SQL injection
	if strings.ContainsAny(name, "'\";--") || strings.ContainsAny(tagLine, "'\";--") {
		log.Printf("SQL injection characters in name or tagLine: name=%s, tagLine=%s", name, tagLine)
		return nil, fmt.Errorf("name or tagLine contains SQL injection characters")
	}

	summonerExists, err := databaseHelper.SummonerExists(name, tagLine, region)
	if err != nil {
		log.Printf("Failed to check if summoner exists: %v", err)
		return nil, fmt.Errorf("failed to check if summoner exists: %v", err)
	}

	if summonerExists {
		log.Printf("Summoner already exists: name=%s, tagLine=%s, region=%s", name, tagLine, region)
	} else {
		summoner, err = apiHelper.GetSummonerByTag(name, tagLine, region)
		if err != nil {
			log.Printf("Failed to fetch summoner data: %v", err)
			return nil, fmt.Errorf("failed to fetch summoner data: %v", err)
		}
		log.Printf("Fetched summoner data: %+v", summoner)

		err = databaseHelper.SaveSummonerToDB(*summoner)
		if err != nil {
			return nil, fmt.Errorf("failed to save summoner to database: %v", err)
		}
	}

	err = databaseHelper.SaveChannelForSummoner(summoner.PUUID, channelID)
	if err != nil {
		return nil, fmt.Errorf("summoner already exists: name=%s, tagLine=%s, region=%s, channekID=%s", name, tagLine, region, channelID)
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
