package commands

import (
	"bytes"
	"main/utils"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/log"
	"github.com/fogleman/gg"
)

var aboutCommand = discord.SlashCommandCreate{
	Name:        "about",
	Description: "About Bored Bot",
}

// Handles the logic for the about command. Sends a message to the user with bot's latency.
func HandleAboutCommand(e *events.ApplicationCommandInteractionCreate) {
	if data := e.SlashCommandInteractionData(); data.CommandName() != "about" {
		return
	}

	message := discord.NewMessageCreateBuilder()

	embed := discord.Embed{
		Title:       "About Bored Bot",
		Description: "Bored Bot is a simple-to-use Discord bot which allows you to get something to do when you're bored! [GitHub repo link.](https://github.com/TisLeo/Bored-Bot)",
		Fields: []discord.EmbedField{
			{
				Name:  "Commands",
				Value: "• `/bored` - get something to do when you're bored. *Some options are missing because they are not valid parameters for BoredAPI.* Leaving options blank returns a random activity!\n• `/ping` - get the bot's latency\n• `/about` - this...\nㅤ\n",
			}, {
				Name:  "Tech Stack",
				Value: "• [Go](https://go.dev/) Language\n• [DisGo](https://github.com/disgoorg/disgo) library\n• [gg](https://github.com/fogleman/gg) graphics library\nㅤ\n",
			},
			{
				Name:  "Help",
				Value: "Use the help channel in the [support server](https://discord.gg/yx3gkv6Why).",
			},
		},
		Color: 0x4B63CF,
		Thumbnail: &discord.EmbedResource{
			URL: "attachment://bored-bot-logo.png",
		},
	}

	message.AddEmbeds(embed)

	if logo, err := getThumbnail(); err != nil {
		log.Errorf("error getting logo for embed thumbnail: %s", err.Error())
	} else {
		reader := bytes.NewReader(logo)
		message.AddFile("bored-bot-logo.png", "Logo of bored bot", reader)
	}

	if err := e.CreateMessage(message.Build()); err != nil {
		log.Errorf("Error responding to slash command '/about': %s", err.Error())
	}
}

// Returns bored bot logo img data
func getThumbnail() ([]byte, error) {
	logo, err := gg.LoadImage("./assets/bored-bot-logo.png")
	if err != nil {
		return nil, err
	}

	if imgData, err := utils.ImgToBytes(logo); err != nil {
		return nil, err
	} else {
		return imgData, nil
	}
}
