package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	databaseHelper "discord-bot/database-helper"
	"discord-bot/features/checkforsummonerupdate"
	"discord-bot/features/offboarding"
	"discord-bot/features/onboarding"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var (
	RemoveCommands = flag.Bool("rmcmd", false, "Remove all commands after shutdowning or not")
)

var s *discordgo.Session

func init() {
	flag.Parse()
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
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
	commands = []*discordgo.ApplicationCommand{
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
		{
			Name:        "ping",
			Description: "Responds with Pong!",
		},
		{
			Name:        "addnumbers",
			Description: "Add two numbers",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "num1",
					Description: "First number",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "num2",
					Description: "Second number",
					Required:    true,
				},
			},
		},
		{
			Name:        "delete",
			Description: "Delete a summoner",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "summoner",
					Description: "Summoner to delete",
					Required:    true,
					Choices:     getSummonerChoices(), // Function to get all summoner choices
				},
			},
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"ping": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Pong!",
				},
			})
		},
		"addnumbers": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			options := i.ApplicationCommandData().Options
			num1 := options[0].IntValue()
			num2 := options[1].IntValue()
			sum := num1 + num2

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("The sum of %d and %d is %d", num1, num2, sum),
				},
			})
		},
		"add": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			options := i.ApplicationCommandData().Options
			name := options[0].StringValue()
			tag := options[1].StringValue()
			summoner, err := onboarding.OnboardSummoner(name, tag)
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
					Content: fmt.Sprintf("Summoner %v with Solo-Rank %v is now registered", summoner.GetNameTag(), summoner.SoloRank.ToString()),
				},
			})
		},
		"delete": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			options := i.ApplicationCommandData().Options
			summonerNameTag := options[0].StringValue()
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

func getSummonerChoices() []*discordgo.ApplicationCommandOptionChoice {
	summoners, err := databaseHelper.LoadSummonersFromFile() // Function to get all summoners
	if err != nil {
		log.Printf("failed to load File: %v", err)
		return []*discordgo.ApplicationCommandOptionChoice{}
	}
	choices := make([]*discordgo.ApplicationCommandOptionChoice, 0, len(summoners))
	for _, summoner := range summoners {
		log.Printf("Summoner: %s, Data: %+v\n", summoner.GetNameTag(), summoner)
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  summoner.GetNameTag(),
			Value: summoner.GetNameTag(),
		})
	}
	return choices
}

func main() {

	// Get the guild ID from the environment variables
	GuildID := os.Getenv("GUILD_ID")
	if GuildID == "" {
		log.Fatal("Guild ID not found in environment variables")
	}

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

	log.Println("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, GuildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
		log.Printf("Command '%v' registered successfully", v.Name)
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

	if *RemoveCommands {
		log.Println("Removing commands...")
		for _, v := range registeredCommands {
			err := s.ApplicationCommandDelete(s.State.User.ID, GuildID, v.ID)
			if err != nil {
				log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
			}
		}
	}

	log.Println("Gracefully shutting down.")
}
