package databaseHelper

import (
	"database/sql"
	"discord-bot/types/match"
	"discord-bot/types/rank"
	"discord-bot/types/summoner"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

var db *sql.DB

// InitDB initializes the database connection with a connection pool
func InitDB() error {
	var err error
	// Construct the PostgreSQL connection string
	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")
	dbHost := os.Getenv("POSTGRES_HOST")

	dataSourceName := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbName)

	db, err = sql.Open("postgres", dataSourceName)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	// Run migrations before opening the database connection
	if err := runMigrations(db); err != nil {
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	db, err = sql.Open("postgres", dataSourceName)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	// Set the maximum number of open connections to the database
	db.SetMaxOpenConns(3)
	// Set the maximum number of idle connections in the pool
	db.SetMaxIdleConns(3)

	return db.Ping()
}

func runMigrations(db *sql.DB) error {
	goose.SetBaseFS(os.DirFS("migrations/db"))

	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("failed to run migrations: %v", err)
	}
	return nil
}

// SaveSummonersToDB saves a map of Summoner instances to the database
func SaveSummonersToDB(summoners map[string]*summoner.Summoner) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	for _, summoner := range summoners {
		_, err := tx.Exec(`
			INSERT INTO Summoner (Name, TagLine, AccountID, ID, PUUID, ProfileIconID, SoloRank, FlexRank, Updated, Region)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			ON CONFLICT (ID) DO UPDATE SET
				Name = EXCLUDED.Name,
				TagLine = EXCLUDED.TagLine,
				AccountID = EXCLUDED.AccountID,
				PUUID = EXCLUDED.PUUID,
				ProfileIconID = EXCLUDED.ProfileIconID,
				SoloRank = EXCLUDED.SoloRank,
				FlexRank = EXCLUDED.FlexRank,
				Updated = EXCLUDED.Updated,
				Region = EXCLUDED.Region
		`, summoner.Name, summoner.TagLine, summoner.AccountID, summoner.ID, summoner.PUUID, summoner.ProfileIconID, summoner.SoloRank, summoner.FlexRank, summoner.Updated, summoner.Region)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to save summoner: %v", err)
		}
	}

	return tx.Commit()
}

// SaveSummonerToDB saves a summoner to the database
func SaveSummonerToDB(summoner summoner.Summoner) error {
	query := `
        INSERT INTO Summoner (Name, TagLine, AccountID, ID, PUUID, ProfileIconID, SoloRank, FlexRank, Updated, Region)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        ON CONFLICT (PUUID) DO UPDATE SET
            Name = EXCLUDED.Name,
            TagLine = EXCLUDED.TagLine,
            AccountID = EXCLUDED.AccountID,
            ID = EXCLUDED.ID,
            ProfileIconID = EXCLUDED.ProfileIconID,
            SoloRank = EXCLUDED.SoloRank,
            FlexRank = EXCLUDED.FlexRank,
            Updated = EXCLUDED.Updated,
            Region = EXCLUDED.Region
    `
	_, err := db.Exec(query, summoner.Name, summoner.TagLine, summoner.AccountID, summoner.ID, summoner.PUUID, summoner.ProfileIconID, summoner.SoloRank, summoner.FlexRank, summoner.Updated, summoner.Region)
	if err != nil {
		return fmt.Errorf("failed to save summoner: %v", err)
	}
	return nil
}

// LoadSummonersFromDB loads a map of Summoner instances from the database
func LoadSummonersFromDB() (map[string]*summoner.Summoner, error) {
	rows, err := db.Query(`SELECT Name, TagLine, AccountID, ID, PUUID, ProfileIconID, SoloRank, FlexRank, Updated, Region FROM Summoner`)
	if err != nil {
		return nil, fmt.Errorf("failed to query summoners: %v", err)
	}
	defer rows.Close()

	summoners := make(map[string]*summoner.Summoner)
	for rows.Next() {
		var s summoner.Summoner
		var soloRank, flexRank int
		err := rows.Scan(&s.Name, &s.TagLine, &s.AccountID, &s.ID, &s.PUUID, &s.ProfileIconID, &soloRank, &flexRank, &s.Updated, &s.Region)
		if err != nil {
			return nil, fmt.Errorf("failed to scan summoner: %v", err)
		}
		s.SoloRank = rank.Rank(soloRank)
		s.FlexRank = rank.Rank(flexRank)
		summoners[s.GetNameTag()] = &s
	}

	return summoners, nil
}

// GetDBSummonerByName retrieves a Summoner instance by name, tag, and region from the database
func GetDBSummonerByName(name, tag, region string) (*summoner.Summoner, error) {
	var s summoner.Summoner
	var soloRank, flexRank int
	err := db.QueryRow(`SELECT Name, TagLine, AccountID, ID, PUUID, ProfileIconID, SoloRank, FlexRank, Updated, Region FROM Summoner WHERE Name = $1 AND TagLine = $2 AND Region = $3`, name, tag, region).Scan(&s.Name, &s.TagLine, &s.AccountID, &s.ID, &s.PUUID, &s.ProfileIconID, &soloRank, &flexRank, &s.Updated, &s.Region)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("summoner with name %s, tag %s, and region %s not found", name, tag, region)
		}
		return nil, fmt.Errorf("failed to get summoner by name, tag, and region: %v", err)
	}
	s.SoloRank = rank.Rank(soloRank)
	s.FlexRank = rank.Rank(flexRank)
	return &s, nil
}

// GetSummonerByPUUID retrieves a Summoner instance by their PUUID from the database
func GetSummonerByPUUIDFromDB(puuid string) (*summoner.Summoner, error) {
	log.Printf("Querying summoner with PUUID: %s", puuid) // Add this line for logging
	var s summoner.Summoner
	var soloRank, flexRank int
	err := db.QueryRow(`SELECT Name, TagLine, AccountID, ID, PUUID, ProfileIconID, SoloRank, FlexRank, Updated, Region FROM Summoner WHERE PUUID = $1`, puuid).Scan(&s.Name, &s.TagLine, &s.AccountID, &s.ID, &s.PUUID, &s.ProfileIconID, &soloRank, &flexRank, &s.Updated, &s.Region)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("summoner with PUUID %s not found", puuid)
		}
		return nil, fmt.Errorf("failed to get summoner by PUUID: %v", err)
	}
	s.SoloRank = rank.Rank(soloRank)
	s.FlexRank = rank.Rank(flexRank)
	return &s, nil
}

// SaveChannelForSummoner saves a channel for a summoner by their PUUID
func SaveChannelForSummoner(puuid, channel string) error {
	_, err := db.Exec(`INSERT INTO SummonerChannel (SummonerPUUID, ChannelID) VALUES ($1, $2) ON CONFLICT DO NOTHING`, puuid, channel)
	if err != nil {
		return fmt.Errorf("failed to save channel for summoner: %v", err)
	}
	return nil
}

// GetChannelsForSummoner retrieves all channels for a summoner by their PUUID
func GetChannelsForSummoner(puuid string) ([]string, error) {
	rows, err := db.Query(`SELECT ChannelID FROM SummonerChannel WHERE SummonerPUUID = $1`, puuid)
	if err != nil {
		return nil, fmt.Errorf("failed to get channels for summoner: %v", err)
	}
	defer rows.Close()

	var channels []string
	for rows.Next() {
		var channel string
		err := rows.Scan(&channel)
		if err != nil {
			return nil, fmt.Errorf("failed to scan channel: %v", err)
		}
		channels = append(channels, channel)
	}

	return channels, nil
}

// DeleteChannelForSummoner deletes a channel for a summoner by their PUUID
func DeleteChannelForSummoner(puuid, channel string) error {
	_, err := db.Exec(`DELETE FROM SummonerChannel WHERE SummonerPUUID = $1 AND ChannelID = $2`, puuid, channel)
	if err != nil {
		return fmt.Errorf("failed to delete channel for summoner: %v", err)
	}
	return nil
}

// SaveOngoingMatchToDB saves an OngoingMatch instance to the database
func SaveOngoingMatchToDB(ongoingMatch *match.Match) error {
	teams, err := json.Marshal(ongoingMatch.Teams)
	if err != nil {
		return fmt.Errorf("failed to marshal teams: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO Match (GameID, GameType, Teams)
		VALUES ($1, $2, $3)
		ON CONFLICT (GameID) DO UPDATE SET
			GameType = EXCLUDED.GameType,
			Teams = EXCLUDED.Teams
	`, ongoingMatch.GameID, ongoingMatch.GameType, teams)
	if err != nil {
		return fmt.Errorf("failed to save ongoing match: %v", err)
	}
	return nil
}

// LoadOngoingMatchFromDB loads an array of Matches instances from the database
func LoadOngoingMatchFromDB() (map[string]*match.Match, error) {
	rows, err := db.Query(`SELECT GameID, GameType, Teams FROM Match`)
	if err != nil {
		return nil, fmt.Errorf("failed to query ongoing matches: %v", err)
	}
	defer rows.Close()

	ongoingMatches := make(map[string]*match.Match)
	for rows.Next() {
		var m match.Match
		var teams []byte
		err := rows.Scan(&m.GameID, &m.GameType, &teams)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ongoing match: %v", err)
		}
		err = json.Unmarshal(teams, &m.Teams)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal teams: %v", err)
		}
		ongoingMatches[m.GameID] = &m
	}

	return ongoingMatches, nil
}

// SummonerExists checks if a summoner with the given name, tag, and region already exists
func SummonerExists(name, tag, region string) (bool, error) {
	var exists bool
	err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM Summoner WHERE Name = $1 AND TagLine = $2 AND Region = $3)`, name, tag, region).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if summoner exists: %v", err)
	}
	return exists, nil
}

// GetOldestSummonerFromDB retrieves the oldest summoner from the database
func GetOldestSummonerFromDB() (*summoner.Summoner, error) {
	row := db.QueryRow(`SELECT Name, TagLine, AccountID, ID, PUUID, ProfileIconID, SoloRank, FlexRank, Updated, Region FROM Summoner ORDER BY Updated ASC LIMIT 1`)

	var name, tagLine, accountID, id, puuid, region string
	var profileIconID int
	var soloRank, flexRank rank.Rank
	var updated time.Time

	err := row.Scan(&name, &tagLine, &accountID, &id, &puuid, &profileIconID, &soloRank, &flexRank, &updated, &region)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no summoners found in the database")
		}
		return nil, fmt.Errorf("failed to retrieve the oldest summoner: %v", err)
	}

	return summoner.NewSummoner(name, tagLine, accountID, id, puuid, profileIconID, soloRank, flexRank, updated, region), nil
}

// IsSummonerMappedToChannel checks if a summoner is mapped to a channel
func IsSummonerMappedToAnyChannel(summonerPUUID string) (bool, error) {
	var exists bool
	query := `
        SELECT EXISTS (
            SELECT 1
            FROM SummonerChannel
            WHERE SummonerPUUID = $1
        )
    `

	err := db.QueryRow(query, summonerPUUID).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return false, fmt.Errorf("failed to check if summoner is mapped to channel: %v", err)
	}

	return exists, nil
}

// IsMatchExists checks if a match already exists in the database by GameID
func IsMatchExists(gameID string) (bool, error) {
	var exists bool
	query := `
        SELECT EXISTS (
            SELECT 1
            FROM Match
            WHERE GameID = $1
        )
    `

	err := db.QueryRow(query, gameID).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return false, fmt.Errorf("failed to check if match exists: %v", err)
	}

	return exists, nil
}

// UpdateSummonerTimestamp updates the Updated timestamp for a given PUUID
func UpdateSummonerTimestamp(puuid string) error {
	query := `
        UPDATE Summoner
        SET Updated = NOW()
        WHERE PUUID = $1
    `

	_, err := db.Exec(query, puuid)
	if err != nil {
		return fmt.Errorf("failed to update timestamp for summoner with PUUID %s: %v", puuid, err)
	}

	return nil
}

// GetOldestSummonerWithChannel retrieves the oldest summoner who also has a channel mapped
func GetOldestSummonerWithChannel() (string, error) {
	var puuid string
	query := `
        SELECT s.PUUID
        FROM Summoner s
        JOIN SummonerChannel sc ON s.PUUID = sc.SummonerPUUID
        ORDER BY s.Updated ASC
        LIMIT 1
    `

	err := db.QueryRow(query).Scan(&puuid)
	if err != nil {
		return "", fmt.Errorf("failed to get the oldest summoner with a channel: %v", err)
	}

	return puuid, nil
}

// UpdateOngoingToFinishedGame updates the entire match, including GameID, GameType, Teams, and Participants
func UpdateOngoingToFinishedGame(oldGameID string, newMatch *match.Match) error {
	teams, err := json.Marshal(newMatch.Teams)
	if err != nil {
		return fmt.Errorf("failed to marshal teams: %v", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	// Update the Match table
	query := `
        UPDATE Match
        SET GameID = $1, GameType = $2, Teams = $3
        WHERE GameID = $4
    `
	_, err = tx.Exec(query, newMatch.GameID, newMatch.GameType, teams, oldGameID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update match with old GameID %s: %v", oldGameID, err)
	}

	// Delete existing participants
	_, err = tx.Exec(`DELETE FROM Participant WHERE GameID = $1`, oldGameID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete participants for match with old GameID %s: %v", oldGameID, err)
	}

	// Insert new participants
	for _, team := range newMatch.Teams {
		for _, participant := range team.Participants {
			perks, err := json.Marshal(participant.Perks)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to marshal perks: %v", err)
			}
			spells, err := json.Marshal(participant.Spells)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to marshal spells: %v", err)
			}
			query := `
                INSERT INTO Participant (GameID, SummonerPUUID, ChampionID, TeamID, Perks, Spells)
                VALUES ($1, $2, $3, $4, $5, $6)
            `
			_, err = tx.Exec(query, newMatch.GameID, participant.Summoner.ID, participant.ChampionID, team.TeamID, perks, spells)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to insert participant for match with GameID %s: %v", newMatch.GameID, err)
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}
