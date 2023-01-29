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

var baseImg, imgErr = gg.LoadImage("./assets/bored-base.png")

const (
	url             = "http://www.boredapi.com/api/activity"
	urlWithKey      = "http://www.boredapi.com/api/activity?key="
	transcriptBtnID = "bored_bot_transcript:" // example structure -> "bored_bot_transcript:42" where 42 is the key
)

type boredActivity struct {
	Activity     string  `json:"activity"`
	Type         string  `json:"type"`
	Participants int     `json:"participants"`
	Price        float32 `json:"price"`
	Key          string  `json:"key"`
}

// Handles the logic for the bored command.
func HandleBoredCommand(e *events.ApplicationCommandInteractionCreate) {
	if data := e.SlashCommandInteractionData(); data.CommandName() != "bored" {
		return
	}

	go func() {
		var messageToSend = discord.MessageCreateBuilder{}

		activity, err := getNewActivity()
		if err != nil {
			log.Errorf("[Bored Command] error getting a new activity: %s", err.Error())
			messageToSend.AddEmbeds(getErrorEmbed())
		} else {
			// button ID is appended with the BoredAPI response's key, used for transcript
			messageToSend.AddActionRow(discord.NewPrimaryButton("Show Transcript", transcriptBtnID+activity.Key))

			if imgData, err := activity.generateImageData(); err != nil {
				messageToSend.AddEmbeds(getImgErrorEmbed())
			} else {
				reader := bytes.NewReader(imgData)
				messageToSend.AddFile("activity.png", "your bored activity", reader)
			}
		}

		if err := e.CreateMessage(messageToSend.Build()); err != nil {
			log.Errorf("Error responding to slash command '/activity': %s", err.Error())
		}
	}()
}

// Returns a new boredActivity.
func getNewActivity() (*boredActivity, error) {
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

// Sends a new GET request to BoredAPI's API with a specific given key. Returns string of data.
func doApiRequestWithKey(key string) (string, error) {
	return doApiRequest(urlWithKey + key)
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

	resp, err := doApiRequestWithKey(key)
	if err != nil {
		log.Errorf("[Bored Command] Error getting transcript activity: %s", err.Error())
		embed = getErrorEmbed()
	} else {
		activity, err := unmarshalResponse(resp)
		if err != nil {
			embed = getErrorEmbed()
		} else {
			embed = activity.getActivityEmbed()
		}
	}

	event.CreateMessage(discord.NewMessageCreateBuilder().SetContent("<@" + event.User().ID.String() + ">").AddEmbeds(embed).Build())
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
