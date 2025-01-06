package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"discord-bot/types"
)

const filename = "summoners.json"

// SaveSummonersToFile saves a map of Summoner instances to a JSON file
func SaveSummonersToFile(newSummoners map[string]*types.Summoner) error {
	// Load existing summoners from the file
	existingSummoners, err := LoadSummonersFromFile()
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to load existing summoners: %v", err)
	}

	// Merge new summoners with existing summoners
	for name, summoner := range newSummoners {
		existingSummoners[name] = summoner
	}

	// Marshal the combined summoners to JSON
	data, err := json.MarshalIndent(existingSummoners, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal summoners: %v", err)
	}

	// Write the combined summoners to the file
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}

// LoadSummonersFromFile loads a map of Summoner instances from a JSON file
func LoadSummonersFromFile() (map[string]*types.Summoner, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	var summoners map[string]*types.Summoner
	err = json.Unmarshal(data, &summoners)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal summoners: %v", err)
	}

	return summoners, nil
}

// GetSummonerByName retrieves a Summoner instance by name from the map
func GetSummonerByName(summoners map[string]*types.Summoner, name string) (*types.Summoner, error) {
	summoner, exists := summoners[name]
	if !exists {
		return nil, fmt.Errorf("summoner with name %s not found", name)
	}
	return summoner, nil
}

func main() {
	// Example usage
	summoners := make(map[string]*types.Summoner)
	summoner := types.NewSummoner("Cedri22c2", "02010", "accountID", "id", "puuid", "Gold IV", "Silver I", time.Now())
	if _, exists := summoners[summoner.Name]; !exists {
		summoners[summoner.Name] = summoner
	}

	// Save the summoners to a file
	err := SaveSummonersToFile(summoners)
	if err != nil {
		fmt.Println("Error saving summoners:", err)
		return
	}

	// Load the summoners from the file
	loadedSummoners, err := LoadSummonersFromFile()
	if err != nil {
		fmt.Println("Error loading summoners:", err)
		return
	}

	// Retrieve a summoner by name
	loadedSummoner, err := GetSummonerByName(loadedSummoners, "Cedri2c2")
	if err != nil {
		fmt.Println("Error retrieving summoner:", err)
		return
	}

	fmt.Printf("Loaded Summoner: %+v\n", loadedSummoner)
}
