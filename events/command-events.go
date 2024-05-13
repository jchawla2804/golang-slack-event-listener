package events

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jchawla2804/golang-slack-event-listener/database"
	"github.com/jchawla2804/golang-slack-event-listener/helper"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"golang.org/x/exp/slices"
)

var (
	cacheClient = database.CreateCache()
	//SlackContext, cancel = context.WithCancel(context.Background())
)

func HandlePlatformInformation(slackClient *slack.Client, businessGroup, businessGroupId string) error {
	cacheClient.Add("business_group_id", businessGroupId, 10*time.Hour)
	cacheClient.Add("business_group_name", businessGroup, 10*time.Hour)

	attachment := slack.Attachment{
		Pretext: "Business Group Information",
		Color:   "#36a64f",
		Text:    "Business Group Name: " + businessGroup + "\nBusiness Group Id: " + businessGroupId,
	}

	_, timestamp, err := slackClient.PostMessage(
		os.Getenv("CHANNEL_ID"),
		slack.MsgOptionAttachments(attachment),
	)

	if err != nil {
		return err
	}

	log.Print("Message Sent at this time " + timestamp)
	return nil
}

func HandleSlackCommands(slackClient *slack.Client, command slack.SlashCommand) error {
	tokenValue, status := cacheClient.Get("access_token")
	if !status {
		log.Println("Please login again")
		return HandleSlackAppMentions(&slackevents.AppMentionEvent{User: command.UserID, Channel: command.ChannelID}, slackClient)
	}
	orgId, status := cacheClient.Get("business_group_id")
	if !status {
		log.Println("Please login again. Org ID not found")
		return HandleSlackAppMentions(&slackevents.AppMentionEvent{User: command.UserID, Channel: command.ChannelID}, slackClient)

	}

	token := tokenValue.(string)

	switch command.Command {
	case "/get-status":

		if !status {
			log.Fatal("No access token is there. Please login")
			return errors.New("Please login again")
		}

		log.Println(token)

		envName := command.Text

		envId, status := cacheClient.Get(envName)
		if !status {
			return errors.New("Please login again. Environeent ID not found for " + envName + " environment")
		}

		appDetails, err := helper.GetAppDetails(token, envId.(string), orgId.(string))
		if err != nil {
			log.Fatal(err)
		}

		slackAttachment := slack.Attachment{
			Text:    appDetails,
			Pretext: fmt.Sprintf("App Details for %s environment ", envName),
		}

		_, _, err = slackClient.PostMessage(os.Getenv("CHANNEL_ID"), slack.MsgOptionAttachments(slackAttachment))
		if err != nil {
			return err
		}

	case "/change-status":

		if !status {
			log.Fatal("No access token is there. Please login")
			return errors.New("Please login again")
		}
		listOfOptions := strings.Split(command.Text, " ")
		envId, status := cacheClient.Get(listOfOptions[1])
		if !status {
			return errors.New("Please login again. Environeent ID not found for " + listOfOptions[1] + " environment")
		}

		log.Println(listOfOptions)
		if !slices.Contains([]string{"stop", "start", "restart"}, listOfOptions[0]) {
			PostMessage(os.Getenv("CHANNEL_ID"), slackClient)
			return nil
		} else {
			_, err := helper.ChangeAppStatus(listOfOptions[0], token, envId.(string), orgId.(string), listOfOptions[2])

			if err != nil {
				log.Print(err.Error())
				return err
			}
			slackAttachment := slack.Attachment{
				Text:    "Status Of API " + listOfOptions[2] + " has changed to " + listOfOptions[0],
				Pretext: "Status has changed",
			}

			_, _, err = slackClient.PostMessage(os.Getenv("CHANNEL_ID"), slack.MsgOptionAttachments(slackAttachment))
			if err != nil {
				return err
			}

		}

	case "/get-asset-info":

		if !status {
			log.Fatal("No access token is there. Please login")
			return errors.New("Please login again")
		}

		response, err := helper.GetAssetInfo(token)
		if err != nil {
			log.Println(err.Error())
			return err
		}

		slackAttachment := slack.Attachment{
			Text:    response,
			Pretext: "Asset Information",
		}

		_, _, err = slackClient.PostMessage(os.Getenv("CHANNEL_ID"), slack.MsgOptionAttachments(slackAttachment))
		if err != nil {
			return err
		}

	case "/list-environments":
		if !status {
			log.Println("No access token is there. Please login")
			return HandleSlackAppMentions(&slackevents.AppMentionEvent{User: command.UserID, Channel: command.ChannelID}, slackClient)
		}
		listOfEnv, err := helper.ListEnvironments(token, orgId.(string))
		if err != nil {
			return err
		}
		var concatenatedString []string
		for _, v := range listOfEnv.Data {
			cacheClient.Add(v.Name, v.ID, 10*time.Hour)
			concatenatedString = append(concatenatedString, fmt.Sprintf("Env-Name : %s\n Env-Id : %s\n Is-Production : %v", v.Name, v.ID, v.IsProduction))
		}

		slackAttachment := slack.Attachment{
			Text:    strings.Join(concatenatedString, "\n\n"),
			Pretext: "List Of Environemnts",
		}

		_, _, err = slackClient.PostMessage(command.ChannelID, slack.MsgOptionAttachments(slackAttachment))
		if err != nil {
			return err
		}
	case "/download-asset":
		if !status {
			log.Fatal("No access token is there. Please login")
			return errors.New("Please login again")
		}
		fileName, err := helper.DownloadAsset(token, orgId.(string), command.Text)
		if err != nil {
			return err
		}
		slackUploadParam := slack.FileUploadParameters{
			Channels: []string{command.ChannelID},
			File:     fileName,
		}

		fileoutput, err := slackClient.UploadFile(slackUploadParam)
		if err != nil {
			log.Printf("Slack Error:- %s", err.Error())
			return err
		}

		log.Printf("Name: %s\n, Url: %s\n", fileoutput.Name, fileoutput.URLPrivate)
		os.Remove(fileName)

	}

	return nil
}

func PostMessage(channelId string, slackClient *slack.Client) {
	attachment := slack.Attachment{
		Pretext: "Please Mention Correct Status",
		Color:   "#36a64f",
		Text:    "Acceptable Status are Start, Stop, Restart",
	}

	_, timestamp, err := slackClient.PostMessage(
		channelId,
		slack.MsgOptionAttachments(attachment),
	)

	if err != nil {
		log.Panic(err)
	}

	log.Print("Message Sent at this time " + timestamp)
}
