package apiHelper

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"discord-bot/internal/app/constants"
	databaseHelper "discord-bot/internal/app/helper/database"
	"discord-bot/internal/logger"
	"discord-bot/types/match"
	"discord-bot/types/rank"
	"discord-bot/types/summoner"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

var (
	client       *http.Client
	requestQueue chan request
)

type request struct {
	url      string
	response chan *http.Response
	err      chan error
}

func init() {
	// Create a custom transport with a shared TLS connection
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // Adjust as needed
		},
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	client = &http.Client{
		Transport: tr,
		Timeout:   30 * time.Second,
	}

	// Initialize the request queue
	requestQueue = make(chan request, 100)

	// Start the request processor
	go processRequests()

	logger.Logger.Info("set RateLimit is: %w per %w equals: %w ms", rateLimiterRequestPerTime.GetMaxTokens(), rateLimiterRequestPerTime.GetInterval(), rateLimiterRequestPerTime.GetInterval()/1000000)
}

func getBaseURL(platform string, region string) (string, error) {
	if baseURL, ok := constants.Platforms[platform]; ok {
		return "https://" + baseURL, nil
	}
	if baseURL, ok := constants.Regions[region]; ok {
		return "https://" + baseURL, nil
	}
	return "", fmt.Errorf("invalid platform or region")
}

func LoadEnv() error {
	if os.Getenv("RIOT_API_TOKEN") != "" {
		return nil
	}
	return godotenv.Load()
}

func waitForRateLimiters() {
	for !rateLimiterPerSecond.Check() || !rateLimiterRequestPerTime.Check() {
		time.Sleep(rateLimiterRequestPerTime.GetInterval())
	}
	rateLimiterRequestPerTime.Allow()
	rateLimiterPerSecond.Allow()
}

func makeRequest(url string) (*http.Response, error) {
	if url == "" {
		return nil, fmt.Errorf("url cannot be empty")
	}

	req := request{
		url:      url,
		response: make(chan *http.Response),
		err:      make(chan error),
	}

	requestQueue <- req

	select {
	case resp := <-req.response:
		return resp, nil
	case err := <-req.err:
		return nil, err
	}
}

func processRequests() {
	for req := range requestQueue {
		waitForRateLimiters()
		resp, err := client.Get(req.url)
		if err != nil {
			logger.Logger.Error("Failed to make request", zap.Error(err))
			req.err <- fmt.Errorf("failed to make request: %w", err)
			continue
		}

		if resp.StatusCode == 404 {
			req.err <- fmt.Errorf("not found")
			continue
		}
		if resp.StatusCode == 429 {
			logger.Logger.Warn("Rate limit exceeded, waiting 20 seconds...")
			time.Sleep(rateLimiterRequestPerTime.GetInterval()*2)
			waitForRateLimiters()
			resp, err = client.Get(req.url)
			if err != nil {
				logger.Logger.Error("Failed to make request after retries", zap.Error(err))
				req.err <- fmt.Errorf("failed to make request after retries: %w", err)
				continue
			}
		}
		if resp.StatusCode != http.StatusOK {
			logger.Logger.Error("Failed to make request", zap.Int("status", resp.StatusCode))
			req.err <- fmt.Errorf("failed to make request: status %d", resp.StatusCode)
			continue
		}

		req.response <- resp
	}
}

func GetSummonerByTag(name, tagLine, region string) (*summoner.Summoner, error) {
	err := LoadEnv()
	if err != nil {
		return nil, fmt.Errorf("error loading .env file")
	}

	apiKey := os.Getenv("RIOT_API_TOKEN")
	if apiKey == "" {
		return nil, fmt.Errorf("API token not found in environment variables")
	}

	baseURL, err := getBaseURL("", "EUROPE")
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/riot/account/v1/accounts/by-riot-id/%s/%s?api_key=%s", baseURL, name, tagLine, apiKey)
	resp, err := makeRequest(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch summoner data: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var accountData struct {
		PUUID   string `json:"puuid"`
		Name    string `json:"gameName"`
		TagLine string `json:"tagLine"`
	}

	err = json.Unmarshal(body, &accountData)
	if err != nil {
		return nil, err
	}

	return GetSummonerByPUUID(accountData.PUUID, region)
}

func GetSummonerProfileIconIDByPUUID(puuid, region string) (int, error) {
	err := LoadEnv()
	if err != nil {
		return 0, fmt.Errorf("error loading .env file")
	}

	apiKey := os.Getenv("RIOT_API_TOKEN")
	if apiKey == "" {
		return 0, fmt.Errorf("API token not found in environment variables")
	}

	baseUrl, err := getBaseURL(region, "")
	if err != nil {
		return 0, err
	}

	url := fmt.Sprintf("%s/lol/summoner/v4/summoners/by-puuid/%s?api_key=%s", baseUrl, puuid, apiKey)
	resp, err := makeRequest(url)
	if err != nil {
		return 0, err
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to fetch summoner data: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var summonerData struct {
		ID            string `json:"id"`
		AccountID     string `json:"accountId"`
		PUUID         string `json:"puuid"`
		ProfileIconID int    `json:"profileIconId"`
		RevisionDate  int64  `json:"revisionDate"`
		SummonerLevel int    `json:"summonerLevel"`
	}

	err = json.Unmarshal(body, &summonerData)
	if err != nil {
		return 0, err
	}

	return summonerData.ProfileIconID, nil
}

func GetSummonerByPUUID(puuid, region string, optionalArgs ...*string) (*summoner.Summoner, error) {
	var name, tag string = "", ""

	if len(optionalArgs) > 0 {
		name = *optionalArgs[0]
	}
	if len(optionalArgs) > 1 {
		tag = *optionalArgs[1]
	}

	err := LoadEnv()
	if err != nil {
		return nil, fmt.Errorf("error loading .env file")
	}

	apiKey := os.Getenv("RIOT_API_TOKEN")
	if apiKey == "" {
		return nil, fmt.Errorf("API token not found in environment variables")
	}

	if name == "" || tag == "" {
		var err error
		name, tag, err = GetNameTagByPUUID(puuid)
		if err != nil {
			return nil, err
		}
	}

	baseUrl, err := getBaseURL(region, "")
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/lol/summoner/v4/summoners/by-puuid/%s?api_key=%s", baseUrl, puuid, apiKey)
	resp, err := makeRequest(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch summoner data: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var summonerData struct {
		ID            string `json:"id"`
		AccountID     string `json:"accountId"`
		PUUID         string `json:"puuid"`
		ProfileIconID int    `json:"profileIconId"`
		RevisionDate  int64  `json:"revisionDate"`
		SummonerLevel int    `json:"summonerLevel"`
	}

	err = json.Unmarshal(body, &summonerData)
	if err != nil {
		return nil, err
	}

	solorank, rankFlex, err := GetSummonerRank(summonerData.ID, region)

	summoner := summoner.NewSummoner(
		name,
		tag, // TagLine
		summonerData.AccountID,
		summonerData.ID,
		puuid,
		summonerData.ProfileIconID,
		solorank,
		rankFlex, // FlexRank
		time.Now(),
		region, // Updated
	)

	if err != nil {
		return summoner, err
	}

	return summoner, nil
}

func GetSummonerRank(summonerID, region string) (rank.Rank, rank.Rank, error) {
	err := LoadEnv()
	if err != nil {
		return 0, 0, fmt.Errorf("error loading .env file")
	}

	apiKey := os.Getenv("RIOT_API_TOKEN")
	if apiKey == "" {
		return 0, 0, fmt.Errorf("API token not found in environment variables")
	}

	baseUrl, err := getBaseURL(region, "")
	if err != nil {
		return 0, 0, err
	}

	url := fmt.Sprintf("%s/lol/league/v4/entries/by-summoner/%s?api_key=%s", baseUrl, summonerID, apiKey)
	resp, err := makeRequest(url)
	if err != nil {
		return 0, 0, err
	}

	if resp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("failed to fetch summoner rank: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

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
		return 0, 0, nil
	}

	var soloRank, flexRank rank.Rank = 0, 0

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

func GetLastRankedMatchIDbyPUUID(puuid string) (string, error) {
	err := LoadEnv()
	if err != nil {
		return "", fmt.Errorf("error loading .env file")
	}

	apiKey := os.Getenv("RIOT_API_TOKEN")
	if apiKey == "" {
		return "", fmt.Errorf("API token not found in environment variables")
	}

	baseUrl, err := getBaseURL("", "EUROPE")
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/lol/match/v5/matches/by-puuid/%s/ids?&start=0&count=1&api_key=%s", baseUrl, puuid, apiKey)
	resp, err := makeRequest(url)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch last ranked match: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

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

func GetMatchByID(matchId string) (*match.Match, error) {
	logger.Logger.Info("GetMatchByID called", zap.String("matchId", matchId)) // Updated code
	err := LoadEnv()
	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}

	parts := strings.Split(matchId, "_")
	region := parts[0]
	logger.Logger.Info("Region extracted from matchId", zap.String("region", region)) // Updated code

	apiKey := os.Getenv("RIOT_API_TOKEN")
	if apiKey == "" {
		return nil, fmt.Errorf("API token not found in environment variables")
	}

	baseUrl, err := getBaseURL("", "EUROPE")
	if err != nil {
		return nil, fmt.Errorf("error getting base URL: %v", err)
	}

	url := fmt.Sprintf("%s/lol/match/v5/matches/%s?api_key=%s", baseUrl, matchId, apiKey)
	logger.Logger.Info("Request URL", zap.String("url", url)) // Updated code
	resp, err := makeRequest(url)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch match data: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	defer resp.Body.Close()

	logger.Logger.Info("Response body", zap.ByteString("body", body)) // Updated code

	var apiResponse struct {
		Metadata struct {
			GameID string `json:"matchId"`
		} `json:"metadata"`
		Info struct {
			QueueID      int `json:"queueId"`
			Participants []struct {
				PUUID      string `json:"puuid"`
				TeamID     int    `json:"teamId"`
				ChampionID int    `json:"championId"`
				Perks      struct {
					StatPerks struct {
						Defense int `json:"defense"`
						Flex    int `json:"flex"`
						Offense int `json:"offense"`
					} `json:"statPerks"`
					Styles []struct {
						Description string `json:"description"`
						Selections  []struct {
							Perk int `json:"perk"`
						} `json:"selections"`
						Style int `json:"style"`
					} `json:"styles"`
				} `json:"perks"`
				ProfileIconID  int    `json:"profileIcon"`
				RiotIdGameName string `json:"riotIdGameName"`
				RiotIdTagLine  string `json:"riotIdTagLine"`
				SummonerID     string `json:"summonerId"`
				Spell1ID       int    `json:"summoner1Id"`
				Spell2ID       int    `json:"summoner2Id"`
				Item0          int    `json:"item0"`
				Item1          int    `json:"item1"`
				Item2          int    `json:"item2"`
				Item3          int    `json:"item3"`
				Item4          int    `json:"item4"`
				Item5          int    `json:"item5"`
				Item6          int    `json:"item6"`
			} `json:"participants"`
		} `json:"info"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	gameType := "Flex"
	if apiResponse.Info.QueueID == 440 {
		gameType = "Flex"
	} else if apiResponse.Info.QueueID == 420 {
		gameType = "Solo/Duo"
	} else {
		gameType = "UNRANKED"
	}

	matchData := &match.Match{
		GameID:   apiResponse.Metadata.GameID,
		Teams:    [2]match.Team{{TeamID: 100}, {TeamID: 200}},
		GameType: gameType,
	}

	for _, participant := range apiResponse.Info.Participants {
		teamIndex := 0
		if participant.TeamID == 200 {
			teamIndex = 1
		}

		var summoner *summoner.Summoner

		summoner, err := databaseHelper.GetSummonerByPUUIDFromDB(participant.PUUID)
		if err != nil {
			logger.Logger.Error("failed to get summoner by PUUID", zap.Error(err)) // Updated code
		}

		if summoner == nil {
			summoner, err = GetSummonerByPUUID(participant.PUUID, region, &participant.RiotIdGameName, &participant.RiotIdTagLine)
			if err != nil {
				logger.Logger.Error("failed to get summoner by PUUID", zap.Error(err)) // Updated code
				continue
			}
		} else {
			summonerIsKnown, err := databaseHelper.IsSummonerMappedToAnyChannel(summoner.PUUID)
			if err != nil {
				logger.Logger.Error("failed to check if summoner is mapped to any channel", zap.Error(err)) // Updated code
			}
			if !summonerIsKnown {
				summoner.SoloRank, summoner.FlexRank, err = GetSummonerRank(summoner.ID, region)
				if err != nil {
					logger.Logger.Error("failed to get summoner rank", zap.Error(err)) // Updated code
				}
			}
		}

		if summoner == nil {
			continue
		}

		// renew Summoner Data to reduce API calls
		summoner.ProfileIconID = participant.ProfileIconID
		summoner.Name = participant.RiotIdGameName
		summoner.TagLine = participant.RiotIdTagLine

		if err != nil {
			return nil, fmt.Errorf("failed to get summoner by PUUID: %v", err)
		}

		var perkIDs []int
		for _, style := range participant.Perks.Styles {
			for _, selection := range style.Selections {
				perkIDs = append(perkIDs, selection.Perk)
			}
		}

		matchData.Teams[teamIndex].Participants = append(matchData.Teams[teamIndex].Participants, match.Participant{
			Summoner: *summoner,
			Perks: match.Perks{
				PerkIDs:      perkIDs,
				PerkStyle:    participant.Perks.Styles[0].Style,
				PerkSubStyle: participant.Perks.Styles[1].Style,
			},
			ChampionID: participant.ChampionID,
			Items: match.Items{
				ItemIDs: []int{participant.Item0, participant.Item1, participant.Item2, participant.Item3, participant.Item4, participant.Item5, participant.Item6},
			},
			Spells: match.Spells{
				SpellIDs: []int{participant.Spell1ID, participant.Spell2ID},
			},
		})
	}

	logger.Logger.Info("Match data", zap.Any("matchData", matchData)) // Updated code
	return matchData, nil
}

func GetOngoingMatchByPUUID(puuid, region string) (*match.Match, error) {
	apiKey := os.Getenv("RIOT_API_TOKEN")
	if apiKey == "" {
		return nil, fmt.Errorf("API token not found in environment variables")
	}

	baseUrl, err := getBaseURL(region, "")
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/lol/spectator/v5/active-games/by-summoner/%s?api_key=%s", baseUrl, puuid, apiKey)
	resp, err := makeRequest(url)
	if err != nil {
		return nil, fmt.Errorf("failed to make request to Riot Games API: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		// No ongoing match found
		return nil, nil
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status: %v", resp.Status)
	}

	var apiResponse struct {
		GameID       int64 `json:"gameId"`
		QueueID      int   `json:"gameQueueConfigId"`
		Participants []struct {
			PUUID      string      `json:"puuid"`
			TeamID     int         `json:"teamId"`
			ChampionID int         `json:"championId"`
			Perks      match.Perks `json:"perks"`
			SummonerID string      `json:"summonerId"`
			Spell1ID   int         `json:"spell1Id"`
			Spell2ID   int         `json:"spell2Id"`
		} `json:"participants"`
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	defer resp.Body.Close()

	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	gameType := "Flex"
	if apiResponse.QueueID == 440 {
		gameType = "Flex"
	} else if apiResponse.QueueID == 420 {
		gameType = "Solo/Duo"
	} else {
		gameType = "UNRANKED"
	}

	gameIDStr := fmt.Sprintf("%d", apiResponse.GameID)

	ongoingMatch := &match.Match{
		GameID:   gameIDStr,
		Teams:    [2]match.Team{{TeamID: 100}, {TeamID: 200}},
		GameType: gameType,
	}

	gameIsKnown, err := databaseHelper.IsMatchExists(gameIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to check if match exists: %v", err)
	}

	if gameIsKnown {
		return nil, fmt.Errorf("match is already known")
	}

	// Populate teams and participants
	for _, participant := range apiResponse.Participants {
		teamIndex := 0
		if participant.TeamID == 200 {
			teamIndex = 1
		}

		summoner, err := databaseHelper.GetSummonerByPUUIDFromDB(participant.PUUID)
		if err != nil {
			logger.Logger.Error("failed to get summoner by PUUID", zap.Error(err)) // Updated code
		}

		if summoner == nil {
			summoner, err = GetSummonerByPUUID(participant.PUUID, region)
			if err != nil {
				logger.Logger.Error("failed to fetch summoner by PUUID", zap.Error(err)) // Updated code
				continue
			} else {
				databaseHelper.SaveSummonerToDB(*summoner)
			}
		} else {
			summoner.SoloRank, summoner.FlexRank, err = GetSummonerRank(summoner.ID, region)
			if err != nil {
				logger.Logger.Error("failed to get new summoner rank", zap.Error(err)) // Updated code
			} else {
				databaseHelper.SaveSummonerToDB(*summoner)
			}
		}

		logger.Logger.Info("Summoner", zap.Any("summoner", summoner)) // Updated code

		ongoingMatch.Teams[teamIndex].Participants = append(ongoingMatch.Teams[teamIndex].Participants, match.Participant{
			Summoner:   *summoner,
			Perks:      participant.Perks,
			ChampionID: participant.ChampionID,
			Items:      match.Items{},
			Spells: match.Spells{
				SpellIDs: []int{participant.Spell1ID, participant.Spell2ID},
			},
		})
	}

	return ongoingMatch, nil
}

func GetNameTagByPUUID(puuid string) (string, string, error) {
	err := LoadEnv()
	if err != nil {
		return "", "", fmt.Errorf("error loading .env file")
	}

	apiKey := os.Getenv("RIOT_API_TOKEN")
	if apiKey == "" {
		return "", "", fmt.Errorf("API token not found in environment variables")
	}

	baseUrl, err := getBaseURL("", "EUROPE")
	if err != nil {
		return "", "", err
	}

	url := fmt.Sprintf("%s/riot/account/v1/accounts/by-puuid/%s?api_key=%s", baseUrl, puuid, apiKey)
	resp, err := makeRequest(url)
	if err != nil {
		return "", "", fmt.Errorf("failed to make request to Riot Games API: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("failed to fetch account data: %s", resp.Status)
	}

	var account struct {
		GameName string `json:"gameName"`
		TagLine  string `json:"tagLine"`
	}

	err = json.NewDecoder(resp.Body).Decode(&account)
	if err != nil {
		return "", "", fmt.Errorf("failed to decode response: %v", err)
	}

	defer resp.Body.Close()

	return account.GameName, account.TagLine, nil
}
