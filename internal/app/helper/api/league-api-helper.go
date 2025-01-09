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

	// Log the response body for debugging
	log.Printf("Response body: %s", string(body))

	var summonerData struct {
		ID            string `json:"id"`
		AccountID     string `json:"accountId"`
		PUUID         string `json:"puuid"`
		ProfileIconID int    `json:"profileIconId"`
		RevisionDate  int64  `json:"revisionDate"`
		SummonerLevel int    `json:"summonerLevel"`
	}

	log.Printf("ProfileIconID: %d", summonerData.ProfileIconID)

	err = json.Unmarshal(body, &summonerData)
	if err != nil {
		return nil, err
	}

	log.Printf("ProfileIconID: %d", summonerData.ProfileIconID)

	solorank, rankFlex, err := GetSummonerRank(summonerData.ID)

	summoner := summoner.NewSummoner(
		name,
		tagLine, // TagLine
		summonerData.AccountID,
		summonerData.ID,
		summonerData.PUUID,
		summonerData.ProfileIconID,
		solorank,
		rankFlex,   // FlexRank
		time.Now(), // Updated
	)

	if err != nil {
		return summoner, err
	}

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

// GetLastRankedMatch fetches the last ranked match of a summoner by their PUUID from the League of Legends API
func GetLastRankedMatch(puuid string) (string, error) {
	err := LoadEnv()
	if err != nil {
		return "", fmt.Errorf("error loading .env file")
	}

	apiKey := os.Getenv("ROPT_API_TOKEN")
	if apiKey == "" {
		return "", fmt.Errorf("API token not found in environment variables")
	}

	url := fmt.Sprintf("https://euw1.api.riotgames.com/lol/match/v5/matches/by-puuid/%s/ids?type=ranked&start=0&count=1&api_key=%s", puuid, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch last ranked match: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var matchIDs []string
	err = json.Unmarshal(body, &matchIDs)
	if err != nil {
		return "", err
	}

	if len(matchIDs) == 0 {
		return "", fmt.Errorf("no ranked matches found for summoner")
	}

	return matchIDs[0], nil
}
