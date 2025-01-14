package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"discord-bot/internal/app/features/checkforsummonerupdate"
	"discord-bot/internal/app/features/offboarding"
	"discord-bot/internal/app/features/onboarding"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var (
	RemoveCommands = flag.Bool("rmcmd", true, "Remove all commands after shutdowning or not")
)

var s *discordgo.Session

func init() {
	flag.Parse()
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	BotToken := os.Getenv("DISCORD_BOT_TOKEN")
	if BotToken == "" {
		log.Fatal("Bot token not found in environment variables")
	}

	GuildID := os.Getenv("GUILD_ID")
	if GuildID == "" {
		log.Fatal("Guild ID not found in environment variables")
	}

	s, err = discordgo.New("Bot " + BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
}

var (
	// commands is a slice of ApplicationCommand pointers that defines the available commands for the Discord bot.
	// The following commands are included:
	// - add: Adds a new summoner with the required options "name" (Ingame Name) and "tag" (Your Riot Tag).
	// - ping: Responds with "Pong!".
	// - delete: Deletes a summoner with the required options "name" (Ingame Name) and "tag" (Your Riot Tag).
	commands = []*discordgo.ApplicationCommand{

		// - add: Adds a new summoner with the required options "name" (Ingame Name) and "tag" (Your Riot Tag).
		{
			Name:        "add",
			Description: "Add a new Summoner",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "name",
					Description: "Ingame Name",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "tag",
					Description: "Your Riot Tag",
					Required:    true,
				},
			},
		},

		// - ping: Responds with "Pong!".
		{
			Name:        "ping",
			Description: "Responds with Pong!",
		},

		// - delete: Deletes a summoner with the required options "name" (Ingame Name) and "tag" (Your Riot Tag).
		{
			Name:        "delete",
			Description: "Delete a summoner",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "name",
					Description: "Ingame Name",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "tag",
					Description: "Your Riot Tag",
					Required:    true,
				},
			},
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"add": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			options := i.ApplicationCommandData().Options
			name := options[0].StringValue()
			tag := options[1].StringValue()
			message, err := onboarding.OnboardSummoner(name, tag)
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf("Failed to onboard summoner: %v", err),
					},
				})
				return
			}
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{message},
				},
			})
		},
		"ping": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Pong!",
				},
			})
		},
		"delete": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			options := i.ApplicationCommandData().Options
			gameName := options[0].StringValue()
			tag := options[1].StringValue()
			summonerNameTag := fmt.Sprintf("%s#%s", gameName, tag)
			log.Printf("Deleting summoner: %v", summonerNameTag)

			err := offboarding.DeleteSummoner(summonerNameTag)
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf("Failed to delete summoner: %v", err),
					},
				})
				return
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Summoner %v has been deleted", summonerNameTag),
				},
			})
		},
	}
)

func removeCommands(s *discordgo.Session, GuildID string) {
	log.Println("Removing commands...")

	// Fetch all existing commands
	commands, err := s.ApplicationCommands(s.State.User.ID, GuildID)
	if err != nil {
		log.Panicf("Cannot fetch commands: %v", err)
	}

	// Delete each command
	for _, v := range commands {
		err := s.ApplicationCommandDelete(s.State.User.ID, GuildID, v.ID)
		log.Printf("Deleting command: %v", v.Name)
		if err != nil {
			log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
		}
	}
}

func addCommands(s *discordgo.Session, GuildID string, commands []*discordgo.ApplicationCommand) error {
	log.Println("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, GuildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
			return err
		}
		registeredCommands[i] = cmd
		log.Printf("Command '%v' registered successfully", v.Name)
	}
	return nil
}

func main() {
	time.Sleep(120 * time.Second)

	// Get the guild ID from the environment variables
	GuildID := os.Getenv("GUILD_ID")
	if GuildID == "" {
		log.Fatal("Guild ID not found in environment variables")
	}

	//s.Debug = true

	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	if *RemoveCommands && GuildID != "" {
		removeCommands(s, GuildID)
	}

	if GuildID != "" {
		addCommands(s, GuildID, commands)
	} else {
		log.Println("Guild ID is not set. Skipping command registration.")
	}

	// Get the channel ID from the environment variables
	ChannelID := os.Getenv("CHANNEL_ID")
	if ChannelID == "" {
		log.Fatal("Channel ID not found in environment variables")
	}

	// Initialize the checkforsummonerupdate package
	checkforsummonerupdate.Initialize(s, ChannelID)

	// Start the rank checking in a separate goroutine
	go checkforsummonerupdate.CheckForUpdates()

	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	if *RemoveCommands && GuildID != "" {
		removeCommands(s, GuildID)
	}

	log.Println("Gracefully shutting down.")
}
