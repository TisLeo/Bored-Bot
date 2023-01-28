package commands

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/log"
)

// Handles the logic for the ping command. Sends a message to the user with bot's latency.
func HandleAboutCommand(e *events.ApplicationCommandInteractionCreate) {
	if data := e.SlashCommandInteractionData(); data.CommandName() != "about" {
		return
	}

	//todo: add logo as thumbnail & discord server for contact link

	embed := discord.Embed{
		Title:       "About Bored Bot",
		Description: "Bored Bot is a simple-to-use Discord bot which allows you to get something to do when you're bored! [GitHub repo link.](https://github.com/TisLeo/Bored-Bot)",
		Fields: []discord.EmbedField{
			{
				Name:  "Commands",
				Value: "• `/activity` - get something to do when you're bored.\n• `/ping` - get the bot's latency\n• `/about` - this...",
			}, {
				Name:  "Tech Stack",
				Value: "• [Go language](https://go.dev/)\n• [disgo](https://github.com/disgoorg/disgo) library\n• [gg](https://github.com/fogleman/gg) graphics library",
			},
			{
				Name:  "Help",
				Value: "Contact the developer in the support forum channel at [tbd]()",
			},
		},
		Color: 0x4B63CF,
	}

	msg := discord.NewMessageCreateBuilder().AddEmbeds(embed).Build()
	if err := e.CreateMessage(msg); err != nil {
		log.Errorf("Error responding to slash command '/about': %s", err.Error())
	}
}
