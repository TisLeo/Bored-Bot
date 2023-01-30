package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"main/utils"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/log"
	"github.com/fogleman/gg"
)

var boredCommand = discord.SlashCommandCreate{
	Name:        "bored",
	Description: "Bored? Get something to do",
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionString{
			Name:        "type",
			Description: "the activity type",
			Required:    false,
			Choices: []discord.ApplicationCommandOptionChoiceString{
				{Name: "education", Value: "education"},
				{Name: "recreational", Value: "recreational"},
				{Name: "social", Value: "social"},
				{Name: "diy", Value: "diy"},
				{Name: "charity", Value: "charity"},
				{Name: "cooking", Value: "cooking"},
				{Name: "relaxation", Value: "relaxation"},
				{Name: "music", Value: "music"},
				{Name: "busywork", Value: "busywork"},
			},
		},
		discord.ApplicationCommandOptionString{
			Name:        "price",
			Description: "the relative price, where 0 is free",
			Required:    false,
			Choices: []discord.ApplicationCommandOptionChoiceString{
				{Name: "0", Value: "0"},
				{Name: "0.1", Value: "0.1"},
				{Name: "0.2", Value: "0.2"},
				{Name: "0.3", Value: "0.3"},
				{Name: "0.4", Value: "0.4"},
				{Name: "0.5", Value: "0.5"},
				{Name: "0.6", Value: "0.6"},
				{Name: "0.8", Value: "0.8"},
			},
		},
		discord.ApplicationCommandOptionString{
			Name:        "participants",
			Description: "the number of participants",
			Required:    false,
			Choices: []discord.ApplicationCommandOptionChoiceString{
				{Name: "1", Value: "1"},
				{Name: "2", Value: "2"},
				{Name: "3", Value: "3"},
				{Name: "4", Value: "4"},
				{Name: "5", Value: "5"},
				{Name: "8", Value: "8"},
			},
		},
	},
}

const (
	url             = "http://www.boredapi.com/api/activity"
	transcriptBtnID = "bored_bot_transcript:" // example structure -> "bored_bot_transcript:42" where 42 is the key
)

type boredActivity struct {
	Error        string  `json:"error"`
	Activity     string  `json:"activity"`
	Type         string  `json:"type"`
	Participants int     `json:"participants"`
	Price        float32 `json:"price"`
	Key          string  `json:"key"`
}

// Handles the logic for the bored command.
func HandleBoredCommand(e *events.ApplicationCommandInteractionCreate) {
	go func() {
		data := e.SlashCommandInteractionData()
		if data.CommandName() != "bored" {
			return
		}

		message := discord.NewMessageCreateBuilder()
		var activity *boredActivity
		var err error

		if len(data.Options) == 0 {
			activity, err = getNewRandomActivity()
		} else {
			activity, err = getActivity(url + getURLQueryFromOpts(data))
		}

		if err != nil {
			log.Errorf("[Bored Command] error getting a new activity: %s", err.Error())
			message.AddEmbeds(getErrorEmbed())
			message.SetEphemeral(true)
		} else if activity.Error != "" {
			message.AddEmbeds(getQueryErrorEmbed())
			message.SetEphemeral(true)
		} else {
			// button ID is appended with the BoredAPI response's key, used for transcript
			message.AddActionRow(discord.NewPrimaryButton("Show Transcript", transcriptBtnID+activity.Key))

			if imgData, err := activity.generateImageData(); err != nil {
				message.AddEmbeds(getImgErrorEmbed())
				message.SetEphemeral(true)
			} else {
				reader := bytes.NewReader(imgData)
				message.AddFile("activity.png", "your bored activity", reader)
			}
		}

		if err := e.CreateMessage(message.Build()); err != nil {
			log.Errorf("Error responding to slash command '/activity': %s", err.Error())
		}
	}()
}

// Returns a structured URL query for BoredAPI, e.g. '?type=cooking&price=0'.
// Should be appended to the global url.
func getURLQueryFromOpts(data discord.SlashCommandInteractionData) string {
	urlQuery := "?"
	for option := range data.Options {
		if value, isPresent := data.OptString(option); isPresent {
			urlQuery += option + "=" + value + "&"
		}
	}

	urlQuery = strings.TrimSuffix(urlQuery, "&")
	return urlQuery
}

// Returns a new random boredActivity.
func getNewRandomActivity() (*boredActivity, error) {
	if activity, err := getActivity(url); err != nil {
		return nil, err
	} else {
		return activity, nil
	}
}

// Returns an activity based on the given url query
func getActivity(url string) (*boredActivity, error) {
	resp, err := doApiRequest(url)
	if err != nil {
		return nil, err
	}

	if activity, err := unmarshalResponse(resp); err != nil {
		return nil, err
	} else {
		return activity, nil
	}
}

// Sends a new GET response from BoredAPI's API. Returns string of data.
func doApiRequest(url string) (string, error) {
	var httpClient = &http.Client{Timeout: 10 * time.Second}
	response, err := httpClient.Get(url)

	if err != nil {
		return "", fmt.Errorf("error with a GET request to '%s': %s", url, err.Error())
	}

	defer response.Body.Close()

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("error parsing the response body: %s", err.Error())
	}

	return string(data), nil
}

// Returns a new boredActivity based on the string of API's json response.
func unmarshalResponse(response string) (*boredActivity, error) {
	activity := boredActivity{}

	if err := json.Unmarshal([]byte(response), &activity); err != nil {
		return nil, fmt.Errorf("error parsing the json: %s", err.Error())
	} else {
		return &activity, nil
	}
}

// Generates the image data (byte array) based on the boredActivity.
// Uses the placeholder image in the assets folder.
func (activity boredActivity) generateImageData() ([]byte, error) {
	baseImg, imgErr := gg.LoadImage("./assets/bored-base.png")
	if imgErr != nil {
		return nil, imgErr
	}

	ctx := gg.NewContextForImage(baseImg)

	ctx.LoadFontFace("./assets/Horta_demo.ttf", 42)
	ctx.SetHexColor("#FFFFFF")

	activity.drawStringsToImg(ctx)

	byteImg, err := utils.ImgToBytes(ctx.Image())
	if err != nil {
		return nil, err
	}
	return byteImg, nil
}

// Adds text (draws strings) about the activity to the given image context
func (activity boredActivity) drawStringsToImg(ctx *gg.Context) {
	words := strings.Split(activity.Activity, " ")
	if len(words) > 5 {
		firstRow := strings.Join(words[0:5], " ")
		secondRow := strings.Join(words[5:], " ")
		ctx.DrawStringAnchored(firstRow, 320, 214, 0.5, 0.5)
		ctx.DrawStringAnchored(secondRow, 320, 262, 0.5, 0.5)
	} else {
		ctx.DrawStringAnchored(strings.Join(words, " "), 320, 236, 0.5, 0.5)
	}

	ctx.DrawStringAnchored(fmt.Sprintf("%.2f", activity.Price), 86, 440, 0.5, 0.5)
	ctx.DrawStringAnchored(activity.Type, 320, 440, 0.5, 0.5)
	ctx.DrawStringAnchored(strconv.Itoa(activity.Participants), 538, 440, 0.5, 0.5)
}

// Sets an interaction response to the 'Get Transcript' button that shows when a user sends a /activity command.
// Sends a new message with the embed based on the BoredAPI response and mentions the user. Uses the button's ID
// (which has the API's key field at the end) to make another API GET request. The key is guaranteed to return the
// same response each time.
func HandleTranscriptButtonResponse(event *events.ComponentInteractionCreate) {
	if !strings.HasPrefix(event.ButtonInteractionData().CustomID(), transcriptBtnID) {
		return
	}

	var embed = discord.Embed{}
	key := event.ButtonInteractionData().CustomID()[21:] // From and excluding the ':' to get the key

	activity, err := getActivity(url + "?key=" + key)
	if err != nil {
		log.Errorf("[Bored Command] Error getting transcript activity: %s", err.Error())
		embed = getErrorEmbed()
	} else {
		embed = activity.getActivityEmbed()
	}

	event.CreateMessage(discord.NewMessageCreateBuilder().SetContent("<@" + event.User().ID.String() + ">").AddEmbeds(embed).SetEphemeral(true).Build())
}

// Returns an embed stating that an error occured getting the activity
func getErrorEmbed() discord.Embed {
	return discord.Embed{
		Title:       "There was an error getting the activity.",
		Description: "If this continues, contact the developer (use `/about`).",
		Color:       0xc93420,
	}
}

// Returns an embed stating that an error occured generating the image
func getImgErrorEmbed() discord.Embed {
	return discord.Embed{
		Title:       "There was an error generating the image.",
		Description: "You can still use the transcript below. If this continues, contact the developer (use `/about`).",
		Color:       0xc93420,
	}
}

// Returns an embed stating that an error occured generating the image
func getQueryErrorEmbed() discord.Embed {
	return discord.Embed{
		Title:       "Oops! No activity exists with your given parameters.",
		Description: "Please try again with different ones. Having no luck? Leave out the options and get a random activity instead!",
		Color:       0xc93420,
	}
}

// Returns an embed with a given activity's values
func (activity boredActivity) getActivityEmbed() discord.Embed {
	embedDesc := fmt.Sprintf("Activity **»** *%s*\n\nType **»** *%s*\n\nRelative Price **»** *%.2f*\n\nParticipants **»** *%d*",
		activity.Activity, activity.Type, activity.Price, activity.Participants)

	return discord.Embed{
		Title:       "BORED? Try this... [Transcript]",
		Description: embedDesc,
		Color:       0x4B63CF,
	}
}
