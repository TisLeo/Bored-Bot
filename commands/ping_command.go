package commands

import (
	"strconv"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/log"
)

var pingCommand = discord.SlashCommandCreate{
	Name:        "ping",
	Description: "Get bot's latency",
}

// Handles the logic for the ping command. Sends a message to the user with bot's latency.
func HandlePingCommand(e *events.ApplicationCommandInteractionCreate) {
	if data := e.SlashCommandInteractionData(); data.CommandName() != "ping" {
		return
	}

	latency := "Pong! (Latency `" + strconv.FormatInt(e.Client().Gateway().Latency().Milliseconds(), 10) + "ms`)"
	embed := discord.Embed{
		Title: latency,
		Color: 0x4bb84b,
	}

	message := discord.NewMessageCreateBuilder().AddEmbeds(embed).Build()
	if err := e.CreateMessage(message); err != nil {
		log.Errorf("Error responding to slash command '/ping': %s", err.Error())
	}
}
