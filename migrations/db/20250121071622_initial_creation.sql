-- +goose Up
-- +goose StatementBegin
CREATE TABLE Summoner (
    Name VARCHAR(255) NOT NULL,
    TagLine VARCHAR(255) NOT NULL,
    AccountID VARCHAR(255) NOT NULL,
    ID VARCHAR(255) NOT NULL,
    PUUID VARCHAR(255) PRIMARY KEY,
    ProfileIconID INT NOT NULL,
    SoloRank VARCHAR(255) NOT NULL,
    FlexRank VARCHAR(255) NOT NULL,
    Updated TIMESTAMP NOT NULL DEFAULT NOW(),
    Region VARCHAR(255) DEFAULT 'euw1'
);

CREATE TABLE SummonerChannel (
    SummonerPUUID VARCHAR(255) NOT NULL,
    ChannelID VARCHAR(255) NOT NULL,
    GuildID VARCHAR(255) NOT NULL,
    PRIMARY KEY (SummonerPUUID, ChannelID),
    FOREIGN KEY (SummonerPUUID) REFERENCES Summoner(PUUID)
);

CREATE TABLE Match (
    GameID VARCHAR(255) NOT NULL,
    GameType VARCHAR(255) NOT NULL,
    Region VARCHAR(255) DEFAULT 'euw1',
    PRIMARY KEY (GameID, Region)
);

CREATE TABLE Participant (
    ParticipantID SERIAL PRIMARY KEY,
    GameID VARCHAR(255) NOT NULL,
    SummonerPUUID VARCHAR(255) NOT NULL,
    ChampionID INT NOT NULL,
    TeamID INT NOT NULL,
    Perks JSONB,
    Spells JSONB,
    FOREIGN KEY (GameID) REFERENCES Match(GameID),
    FOREIGN KEY (SummonerPUUID) REFERENCES Summoner(PUUID),
    UNIQUE (GameID, SummonerPUUID) -- Add this line to create the unique constraint
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS Participant;
DROP TABLE IF EXISTS Match;
DROP TABLE IF EXISTS SummonerChannel;
DROP TABLE IF EXISTS Summoner;
-- +goose StatementEnd
