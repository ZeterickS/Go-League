package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"time"

	"discord-bot/internal/app/constants"
	"discord-bot/internal/app/features/checkforsummonerupdate"
	"discord-bot/internal/app/features/offboarding"
	"discord-bot/internal/app/features/onboarding"
	databaseHelper "discord-bot/internal/app/helper/database"
	"discord-bot/internal/logger"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

var (
	RemoveCommands = flag.Bool("rmcmd", true, "Remove all commands after shutdowning or not")
)

var s *discordgo.Session

func init() {
	flag.Parse()
	err := godotenv.Load()
	if err != nil {
		logger.Logger.Warn("Error loading .env file")
	}

	BotToken := os.Getenv("DISCORD_BOT_TOKEN")
	if BotToken == "" {
		logger.Logger.Fatal("Bot token not found in environment variables")
	}

	s, err = discordgo.New("Bot " + BotToken)
	if err != nil {
		logger.Logger.Fatal("Invalid bot parameters", zap.Error(err))
	}

	debug := os.Getenv("DEBUG")
	if debug == "true" {
		go func() {
			logger.Logger.Info("Starting pprof server on :6060")
			logger.Logger.Error(http.ListenAndServe("0.0.0.0:6060", nil).Error())
		}()
	}
}

var (
	// commands is a slice of ApplicationCommand pointers that defines the available commands for the Discord bot.
	// The following commands are included:
	// - add: Adds a new summoner with the required options "name" (Ingame Name) and "tag" (Your Riot Tag).
	// - ping: Responds with "Pong!".
	// - delete: Deletes a summoner with the required options "name" (Ingame Name) and "tag" (Your Riot Tag).
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "add",
			Description: "Add a new Summoner",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "region",
					Description: "Your League Region",
					Required:    true,
					Choices: func() []*discordgo.ApplicationCommandOptionChoice {
						choices := []*discordgo.ApplicationCommandOptionChoice{}
						for _, key := range constants.GetPlatformKeys() {
							choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
								Name:  key,
								Value: key,
							})
						}
						return choices
					}(),
				},
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
			region := options[0].StringValue()
			name := options[1].StringValue()
			tag := options[2].StringValue()
			go func() {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Adding " + name + " " + tag + "... Waiting for RIOT API. Depending on the server load, this may take a while.",
					},
				})
			}()
			message, err := onboarding.OnboardSummoner(name, tag, region, i.ChannelID)
			if err != nil {
				errormessage := fmt.Sprintf("Failed to onboard summoner: %v", err)
				s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: &errormessage,
				})
				return
			}
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: new(string),
				Embeds:  &[]*discordgo.MessageEmbed{message},
			})
		},
		"ping": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			logger.Logger.Debug("Ping command received")
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
			logger.Logger.Info("Deleting summoner", zap.String("summoner", summonerNameTag))

			err := offboarding.DeleteSummoner(summonerNameTag, i.ChannelID)
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

func removeCommands(s *discordgo.Session) {
	logger.Logger.Info("Removing commands...")

	for _, Guild := range s.State.Guilds {
		logger.Logger.Info("Removing commands for guild", zap.String("guildID", Guild.ID))

		// Fetch all existing commands
		commands, err := s.ApplicationCommands(s.State.User.ID, Guild.ID)
		if err != nil {
			logger.Logger.Error("Cannot fetch commands", zap.Error(err))
			continue
		}
		for _, cmd := range commands {
			logger.Logger.Debug("Existing command", zap.String("commandName", cmd.Name))
		}

		// Delete each command
		for _, v := range commands {
			err := s.ApplicationCommandDelete(s.State.User.ID, Guild.ID, v.ID)
			logger.Logger.Debug("Deleting command", zap.String("commandName", v.Name))
			if err != nil {
				logger.Logger.Error("Cannot delete command", zap.String("commandName", v.Name), zap.Error(err))
			}
		}
	}
}

func addCommandsIfNotRegistered(s *discordgo.Session, commands []*discordgo.ApplicationCommand) error {
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
	for _, guild := range s.State.Guilds {
		logger.Logger.Info("Registering commands for guild", zap.String("guildID", guild.ID))
		existingCommands, err := s.ApplicationCommands(s.State.User.ID, guild.ID)
		if err != nil {
			return err
		}

		existingCommandNames := make(map[string]bool)
		for _, cmd := range existingCommands {
			existingCommandNames[cmd.Name] = true
		}

		for _, cmd := range commands {
			if !existingCommandNames[cmd.Name] {
				_, err := s.ApplicationCommandCreate(s.State.User.ID, guild.ID, cmd)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func main() {
	logger.InitLogger()
	logger.Logger.Info("Starting application... Waiting 60s to wait for Database and RateLimit to clear")

	time.Sleep(60 * time.Second)
	logger.Logger.Info("Woke up after initial sleep")

	err := databaseHelper.InitDB()
	if err != nil {
		logger.Logger.Fatal("Failed to initialize database", zap.Error(err))
	}
	logger.Logger.Info("Database initialized successfully")

	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		logger.Logger.Info("Logged in as", zap.String("username", s.State.User.Username), zap.String("discriminator", s.State.User.Discriminator))
	})

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	err = s.Open()
	if err != nil {
		logger.Logger.Fatal("Cannot open the session", zap.Error(err))
	}

	removeCommands(s)

	addCommandsIfNotRegistered(s, commands)

	// Initialize the checkforsummonerupdate package
	logger.Logger.Info("Initializing checkforsummonerupdate package")
	checkforsummonerupdate.Initialize(s)

	// Start the rank checking in a separate goroutine
	logger.Logger.Info("Starting rank checking goroutine")
	go checkforsummonerupdate.CheckForUpdates()

	defer func() {
		logger.Logger.Info("Closing session")
		s.Close()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	logger.Logger.Info("Press Ctrl+C to exit")
	<-stop
	logger.Logger.Info("Application exiting")
}
