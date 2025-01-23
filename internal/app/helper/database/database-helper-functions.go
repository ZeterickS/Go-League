package databaseHelper

import (
	"database/sql"
	"discord-bot/types/match"
	"discord-bot/types/rank"
	"discord-bot/types/summoner"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"discord-bot/internal/logger"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
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
		logger.Logger.Fatal("Failed to connect to the database", zap.Error(err))
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
func GetDBSummonerByName(name, tag string) (*summoner.Summoner, error) {
	var s summoner.Summoner
	var soloRank, flexRank int
	err := db.QueryRow(`SELECT Name, TagLine, AccountID, ID, PUUID, ProfileIconID, SoloRank, FlexRank, Updated, Region FROM Summoner WHERE LOWER(Name) = LOWER($1) AND LOWER(TagLine) = LOWER($2)`, name, tag).Scan(&s.Name, &s.TagLine, &s.AccountID, &s.ID, &s.PUUID, &s.ProfileIconID, &soloRank, &flexRank, &s.Updated, &s.Region)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("summoner with name %s, tag %s not found", name, tag)
		}
		return nil, fmt.Errorf("failed to get summoner by name, tag, and region: %v", err)
	}
	s.SoloRank = rank.Rank(soloRank)
	s.FlexRank = rank.Rank(flexRank)
	return &s, nil
}

// GetSummonerByPUUID retrieves a Summoner instance by their PUUID from the database
func GetSummonerByPUUIDFromDB(puuid string) (*summoner.Summoner, error) {
	logger.Logger.Info("Querying summoner with PUUID", zap.String("PUUID", puuid))
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
func SaveChannelForSummoner(puuid, channel, guildID string) error {
	res, err := db.Exec(`INSERT INTO SummonerChannel (SummonerPUUID, ChannelID, GuildID) VALUES ($1, $2, $3) ON CONFLICT (SummonerPUUID, ChannelID) DO NOTHING`, puuid, channel, guildID)
	if err != nil {
		return fmt.Errorf("failed to save channel for summoner: %v", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("channel for summoner already exists")
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

// DeleteChannelForSummonerByName deletes a channel for a summoner by their name, tag, and region
func DeleteChannelForSummonerByName(name, tag, channel string) error {
	summoner, err := GetDBSummonerByName(name, tag)
	if err != nil {
		return fmt.Errorf("failed to get summoner by name, tag: %v", err)
	}

	err = DeleteChannelForSummoner(summoner.PUUID, channel)
	if err != nil {
		return fmt.Errorf("failed to delete channel for summoner: %v", err)
	}

	return nil
}

// DeleteChannelForSummoner deletes a channel for a summoner by their PUUID
func DeleteChannelForSummoner(puuid, channel string) error {
	res, err := db.Exec(`DELETE FROM SummonerChannel WHERE SummonerPUUID = $1 AND ChannelID = $2`, puuid, channel)
	if err != nil {
		return fmt.Errorf("failed to delete channel for summoner: %v", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("summoner with PUUID %s and channel %s not known", puuid, channel)
	}
	return nil
}

// DeleteChannel deletes a channel by its ID
func DeleteChannel(channelID string) error {
	res, err := db.Exec(`DELETE FROM SummonerChannel WHERE ChannelID = $1`, channelID)
	if err != nil {
		return fmt.Errorf("failed to delete channel: %v", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("channel with ID %s not known", channelID)
	}
	return nil
}

// DeleteGuild deletes a guild by its ID
func DeleteGuild(guildID string) error {
	res, err := db.Exec(`DELETE FROM SummonerChannel WHERE GuildID = $1`, guildID)
	if err != nil {
		return fmt.Errorf("failed to delete guild: %v", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("guild with ID %s not known", guildID)
	}
	return nil
}

// SaveOngoingMatchToDB saves an OngoingMatch instance to the database
func SaveOngoingMatchToDB(ongoingMatch *match.Match) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	_, err = tx.Exec(`
        INSERT INTO Match (GameID, GameType)
        VALUES ($1, $2)
        ON CONFLICT (GameID) DO UPDATE SET
            GameType = EXCLUDED.GameType
    `, ongoingMatch.GameID, ongoingMatch.GameType)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to save ongoing match: %v", err)
	}

	for _, team := range ongoingMatch.Teams {
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
			_, err = tx.Exec(`
                INSERT INTO Participant (GameID, SummonerPUUID, ChampionID, TeamID, Perks, Spells)
                VALUES ($1, $2, $3, $4, $5, $6)
                ON CONFLICT (GameID, SummonerPUUID) DO UPDATE SET
                    ChampionID = EXCLUDED.ChampionID,
                    TeamID = EXCLUDED.TeamID,
                    Perks = EXCLUDED.Perks,
                    Spells = EXCLUDED.Spells
            `, ongoingMatch.GameID, participant.Summoner.PUUID, participant.ChampionID, team.TeamID, perks, spells)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to save participant: %v", err)
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// LoadOngoingMatchFromDB loads an array of Matches instances from the database
func LoadOngoingMatchFromDB() (map[string]*match.Match, error) {
	rows, err := db.Query(`SELECT GameID, GameType FROM Match`)
	if err != nil {
		return nil, fmt.Errorf("failed to query ongoing matches: %v", err)
	}
	defer rows.Close()

	ongoingMatches := make(map[string]*match.Match)
	for rows.Next() {
		var m match.Match
		err := rows.Scan(&m.GameID, &m.GameType)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ongoing match: %v", err)
		}

		// Load participants for the match
		participantRows, err := db.Query(`SELECT SummonerPUUID, ChampionID, TeamID, Perks, Spells FROM Participant WHERE GameID = $1`, m.GameID)
		if err != nil {
			return nil, fmt.Errorf("failed to query participants: %v", err)
		}
		defer participantRows.Close()

		for participantRows.Next() {
			var p match.Participant
			var perks, spells []byte
			var teamId int
			err := participantRows.Scan(&p.Summoner.PUUID, &p.ChampionID, &teamId, &perks, &spells)
			if err != nil {
				return nil, fmt.Errorf("failed to scan participant: %v", err)
			}
			err = json.Unmarshal(perks, &p.Perks)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal perks: %v", err)
			}
			err = json.Unmarshal(spells, &p.Spells)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal spells: %v", err)
			}
			if teamId == 100 {
				m.Teams[0].Participants = append(m.Teams[0].Participants, p)
			} else if teamId == 200 {
				m.Teams[1].Participants = append(m.Teams[1].Participants, p)
			} else if teamId == 0 {
				m.Teams[0].Participants = append(m.Teams[0].Participants, p)
			} else if teamId == 1 {
				m.Teams[1].Participants = append(m.Teams[1].Participants, p)
			}
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

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	// Delete existing participants
	_, err = tx.Exec(`DELETE FROM Participant WHERE GameID = $1`, oldGameID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete participants for match with old GameID %s: %v", oldGameID, err)
	}

	// Update the Match table
	query := `
        UPDATE Match
        SET GameID = $1, GameType = $2
        WHERE GameID = $3
    `
	_, err = tx.Exec(query, newMatch.GameID, newMatch.GameType, oldGameID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update match with old GameID %s: %v", oldGameID, err)
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
