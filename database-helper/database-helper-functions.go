package databaseHelper

import (
	"discord-bot/types/summoner"
	"encoding/json"
	"fmt"
	"os"
)

const filename = "summoners.json"

// SaveSummonersToFile saves a map of Summoner instances to a JSON file
func SaveSummonersToFile(summoners map[string]*summoner.Summoner) error {
	// Marshal the summoners to JSON
	data, err := json.MarshalIndent(summoners, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal summoners: %v", err)
	}

	// Write the summoners to the file
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}

// LoadSummonersFromFile loads a map of Summoner instances from a JSON file
func LoadSummonersFromFile() (map[string]*summoner.Summoner, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]*summoner.Summoner), nil
		}
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	if len(data) == 0 {
		return make(map[string]*summoner.Summoner), nil
	}

	var summoners map[string]*summoner.Summoner
	err = json.Unmarshal(data, &summoners)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal summoners: %v", err)
	}

	return summoners, nil
}

// GetSummonerByName retrieves a Summoner instance by name from the map
func GetSummonerByName(summoners map[string]*summoner.Summoner, name string) (*summoner.Summoner, error) {
	summoner, exists := summoners[name]
	if !exists {
		return nil, fmt.Errorf("summoner with name %s not found", name)
	}
	return summoner, nil
}
