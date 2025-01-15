# Discord Bot for League of Legends

This project is a Discord bot written in Go that interacts with the Riot Games API to provide information about active games and summoner updates. It uses the `discordgo` library to interact with Discord and `github.com/bwmarrin/discordgo` for Discord API interactions.

## Project Structure

- **Main Application**: Contains the main application logic.
- **API Helper**: Functions to interact with external APIs.
- **Discord Bot**: Integration with Discord for sending messages and notifications.
- **Database Helper**: Functions to save and load data.
- **CI/CD**: Continuous Integration and Deployment setup using GitHub Actions and a helper script.

## Getting Started

### Prerequisites

- Go 1.23.4 or higher
- Docker (for CI/CD)
- GitHub CLI (for managing GitHub repositories)

### Installation

1. Create a [.env](http://_vscodecontentref_/2) file in the root directory of the project and add your Riot Games API key and Discord bot token:
    ```env
    DISCORD_BOT_TOKEN="your-discord-bot-token"
    ROPT_API_TOKEN="your-riot-api-key"
    CHANNEL_ID="your-discord-channel-id"
    GUILD_ID="your-discord-guild-id"
    API_RATE_LIMIT_2_MINUTE=100
    API_RATE_LIMIT_SECOND=20
    GITHUB_TOKEN="your-github-token"
    GITHUB_USERNAME="your-github-username"
    DEVELOPMENT=False
    ```

2. Make the [pull_and_run.sh](http://_vscodecontentref_/3) script executable:
    ```sh
    chmod +x pull_and_run.sh
    ```

3. Set up a cron job to run the [pull_and_run.sh](http://_vscodecontentref_/4) script periodically:
    ```sh
    crontab -e
    ```

    Add the following line to run the script every 5 minutes:
    ```sh
    */5 * * * * /path/to/your/repository/pull_and_run.sh
    ```

### Running the Bot

To run the bot, use the following command:
```sh
go run main.go