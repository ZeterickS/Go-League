package apiHelper

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"discord-bot/types/rank"
	"discord-bot/types/summoner"

	"github.com/joho/godotenv"
)

// LoadEnv loads environment variables from a .env file
func LoadEnv() error {
	if os.Getenv("ROPT_API_TOKEN") != "" {
		return nil
	}
	return godotenv.Load()
}

// GetSummonerByTag fetches summoner data by tag from the League of Legends API
func GetSummonerByTag(name, tagLine string) (*summoner.Summoner, error) {
	err := LoadEnv()
	if err != nil {
		return nil, fmt.Errorf("error loading .env file")
	}

	apiKey := os.Getenv("ROPT_API_TOKEN")
	if apiKey == "" {
		return nil, fmt.Errorf("API token not found in environment variables")
	}

	url := fmt.Sprintf("https://europe.api.riotgames.com/riot/account/v1/accounts/by-riot-id/%s/%s?api_key=%s", name, tagLine, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch summoner data: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var accountData struct {
		PUUID string `json:"puuid"`
	}

	err = json.Unmarshal(body, &accountData)
	if err != nil {
		return nil, err
	}

	return GetSummonerByPUUID(accountData.PUUID, name, tagLine)
}

// GetSummonerByPUUID fetches summoner data by PUUID from the League of Legends API
func GetSummonerByPUUID(puuid, name, tagLine string) (*summoner.Summoner, error) {
	err := LoadEnv()
	if err != nil {
		return nil, fmt.Errorf("error loading .env file")
	}

	apiKey := os.Getenv("ROPT_API_TOKEN")
	if apiKey == "" {
		return nil, fmt.Errorf("API token not found in environment variables")
	}

	url := fmt.Sprintf("https://euw1.api.riotgames.com/lol/summoner/v4/summoners/by-puuid/%s?api_key=%s", puuid, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch summoner data: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var summonerData struct {
		ID            string `json:"id"`
		AccountID     string `json:"accountId"`
		PUUID         string `json:"puuid"`
		Name          string `json:"name"`
		ProfileIconID int    `json:"profileIconId"`
		RevisionDate  int64  `json:"revisionDate"`
		SummonerLevel int    `json:"summonerLevel"`
	}

	err = json.Unmarshal(body, &summonerData)
	if err != nil {
		return nil, err
	}

	solorank, rankFlex, err := GetSummonerRank(summonerData.ID)
	if err != nil {
		return nil, err
	}

	summoner := summoner.NewSummoner(
		name,
		tagLine, // TagLine
		summonerData.AccountID,
		summonerData.ID,
		summonerData.PUUID,
		solorank,
		0,          // LastSoloRank
		rankFlex,   // FlexRank
		0,          // LastFlexRank
		time.Now(), // Updated
	)
	return summoner, nil
}

// GetSummonerRank fetches the rank and division of a summoner by their ID from the League of Legends API
func GetSummonerRank(summonerID string) (rank.Rank, rank.Rank, error) {
	err := LoadEnv()
	if err != nil {
		return 0, 0, fmt.Errorf("error loading .env file")
	}

	apiKey := os.Getenv("ROPT_API_TOKEN")
	if apiKey == "" {
		return 0, 0, fmt.Errorf("API token not found in environment variables")
	}

	url := fmt.Sprintf("https://euw1.api.riotgames.com/lol/league/v4/entries/by-summoner/%s?api_key=%s", summonerID, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("failed to fetch summoner rank: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, err
	}

	// Log the response body for debugging
	log.Printf("Response body: %s", string(body))

	var rankData []struct {
		QueueType    string `json:"queueType"`
		Tier         string `json:"tier"`
		Rank         string `json:"rank"`
		LeaguePoints int    `json:"leaguePoints"`
	}

	err = json.Unmarshal(body, &rankData)
	if err != nil {
		return 0, 0, err
	}

	if len(rankData) == 0 {
		return 0, 0, fmt.Errorf("no rank data found for summoner")
	}

	var soloRank, flexRank rank.Rank

	for _, entry := range rankData {
		rankStr := fmt.Sprintf("%s %s %d LP", entry.Tier, entry.Rank, entry.LeaguePoints)
		if entry.QueueType == "RANKED_SOLO_5x5" {
			soloRank = rank.FromString(rankStr)
		} else if entry.QueueType == "RANKED_FLEX_SR" {
			flexRank = rank.FromString(rankStr)
		}
	}

	return soloRank, flexRank, nil
}
