package main

import (
	"context"
	"main/commands"
	"os"
	"os/signal"
	"syscall"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/log"
	"github.com/joho/godotenv"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Info("Loading Bored Bot...")

	log.Info("Loading .env file...")
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env: ", err)
	}

	// Create client. Add intents and event listeners
	log.Info("Loading bot client and handlers...")
	client, err := disgo.New(os.Getenv("BORED_BOT_TOKEN"),
		bot.WithGatewayConfigOpts(
			gateway.WithIntents(gateway.IntentGuilds),
			gateway.WithPresence(gateway.NewWatchingPresence("Bored people", discord.OnlineStatusOnline, false)),
		),
		bot.WithEventListenerFunc(commands.HandlePingCommand),
		bot.WithEventListenerFunc(commands.HandleBoredCommand),
		bot.WithEventListenerFunc(commands.HandleTranscriptButtonResponse),
		bot.WithEventListenerFunc(commands.HandleAboutCommand),
		bot.WithEventListenerFunc(func(e *events.GuildsReady) {
			log.Infof("Bot currently in: %d server(s)", e.Client().Caches().Guilds().Len())
		}),
		bot.WithCacheConfigOpts(
			cache.WithCacheFlags(cache.FlagGuilds),
		),
	)
	if err != nil {
		log.Fatal("Error while building disgo: ", err)
	}

	// Shutdown logic
	defer func() {
		client.Close(context.TODO())
		log.Info("Shutting down Bored Bot...")
	}()

	// Register slash commands
	log.Info("Registering slash commands...")
	if _, err := client.Rest().SetGlobalCommands(client.ApplicationID(), commands.Commands); err != nil {
		log.Fatal("Error registering slash commands: ", err)
	}

	// Open gateway
	log.Info("Opening gateway...")
	if err := client.OpenGateway(context.TODO()); err != nil {
		log.Fatal("Error connecting to gateway: ", err)
	}

	log.Info(`
	_____               _    _____     _      _____           _     
	| __  |___ ___ ___ _| |  | __  |___| |_   | __  |___ ___ _| |_ _ 
	| __ -| . |  _| -_| . |  | __ -| . |  _|  |    -| -_| .'| . | | |
	|_____|___|_| |___|___|  |_____|___|_|    |__|__|___|__,|___|_  |
	                                                            |___|
	` + "\n")

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-s
}
