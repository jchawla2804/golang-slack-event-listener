package events

import (
	"encoding/json"
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

func HandleSlackEventMessage(event slackevents.EventsAPIEvent, slackClient *slack.Client) error {
	switch event.Type {
	case slackevents.CallbackEvent:
		innerEvent := event.InnerEvent

		switch ev := innerEvent.Data.(type) {

		case *slackevents.AppMentionEvent:
			log.Println("Bot Id ", ev.BotID)
			err := HandleSlackAppMentions(ev, slackClient)
			if err != nil {
				return err
			}
		}
	default:
		return errors.New("unsupported event type")
	}
	return nil
}

func HandleSlackCommands(slackClient *slack.Client, command slack.SlashCommand) error {
	tokenValue, status := cacheClient.Get("access_token")

	token := tokenValue.(string)

	switch command.Command {
	case "/get-status":

		if !status {
			log.Fatal("No access token is there. Please login")
			return errors.New("Please login again")
		}

		log.Println(token)

		envName := command.Text
		appDetails, err := helper.GetAppDetails(token, envName)
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
		log.Println(listOfOptions)
		if !slices.Contains([]string{"stop", "start", "restart"}, listOfOptions[0]) {
			PostMessage(os.Getenv("CHANNEL_ID"), slackClient)
			return nil
		} else {
			_, err := helper.ChangeAppStatus(listOfOptions[0], token, listOfOptions[2], listOfOptions[1])

			if err != nil {
				log.Print(err.Error())
				return err
			}
			slackAttachment := slack.Attachment{
				Text:    "Status Of API " + listOfOptions[1] + " has changed to " + listOfOptions[0],
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
			log.Fatal("No access token is there. Please login")
			return errors.New("Please login again")
		}
		output, err := helper.ListEnvironments(token)
		if err != nil {
			return err
		}

		slackAttachment := slack.Attachment{
			Text:    output,
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
		fileName, err := helper.DownloadAsset(token, command.Text)
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

func HandleSlackAppMentions(appMentionEvent *slackevents.AppMentionEvent, slackClient *slack.Client) error {
	slackUser, err := slackClient.GetUserInfo(appMentionEvent.User)
	if err != nil {
		log.Println("Err 1")
		log.Print(err.Error())
		return err
	}

	text := strings.ToLower(appMentionEvent.Text)
	slackAttachment := slack.Attachment{}

	buttonBlockElement1 := slack.NewButtonBlockElement("button1", "basic-auth", &slack.TextBlockObject{Type: slack.PlainTextType, Text: "Login"})

	blockSet := []slack.Block{
		//sectionObject,
		slack.NewSectionBlock(&slack.TextBlockObject{Type: slack.MarkdownType, Text: "PLease Choose login option"}, nil, nil),
		slack.NewActionBlock("actionblock789", buttonBlockElement1),
	}

	if strings.Contains(text, "hello") {
		slackAttachment.Text = "Welcome To MuleSoft slack bot "
		slackAttachment.Color = "#4af030"
		//slackAttachment.Pretext = fmt.Sprint("Greetings")

	} else {
		slackAttachment.Text = fmt.Sprintf("How Can I help You %s", slackUser.Name)
		slackAttachment.Pretext = "How Can I help you ?"
		//slackAttachment.Color = "#3d3d3d"
	}

	b, err := json.Marshal(blockSet)
	if err != nil {
		log.Println(err.Error())
	}

	log.Println(string(b))

	_, timestamp, err := slackClient.PostMessage(os.Getenv("CHANNEL_ID"), slack.MsgOptionBlocks(blockSet...))

	if err != nil {
		log.Println("Error Happened While sending message")
		log.Print(err.Error())
		return err
	}
	fmt.Printf("Message Was Sent at this time %s", timestamp)
	return nil

}

func HandleInteractiveDialogBoxEvent(slackClient *slack.Client, buttonValue, triggerId string) error {
	modal := slack.ModalViewRequest{}
	modal.Title = &slack.TextBlockObject{Type: slack.PlainTextType, Text: "Login To Platform"}
	modal.Submit = &slack.TextBlockObject{Type: slack.PlainTextType, Text: "Submit"}
	modal.Type = slack.VTModal

	blockSet := []slack.Block{
		slack.NewInputBlock("username", &slack.TextBlockObject{Type: slack.PlainTextType, Text: "Enter Username"}, nil, slack.PlainTextInputBlockElement{Type: slack.METPlainTextInput, ActionID: "user"}),
		slack.NewInputBlock("password", &slack.TextBlockObject{Type: slack.PlainTextType, Text: "Enter Password"}, nil, slack.PlainTextInputBlockElement{Type: slack.METPlainTextInput, ActionID: "pass"}),
	}

	modal.Blocks = slack.Blocks{BlockSet: blockSet}

	//_, timestamp, err := slackClient.PostMessage(os.Getenv("CHANNEL_ID"), slack)

	data, err := slackClient.OpenView(triggerId, modal)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	log.Println(data)
	return nil

}

func HandleLogin(slackClient *slack.Client, username, password string) error {
	token, err := helper.GetToken(username, password)
	if err != nil {
		slackAttachment := slack.Attachment{
			Text:    "ERROR",
			Pretext: "Invalid username/password",
		}
		slackClient.PostMessage(os.Getenv("CHANNEL_ID"), slack.MsgOptionAttachments(slackAttachment))
		return errors.New("unable To login")
	}

	platformDetails, err := helper.GetPlatformInformation(token.(string))
	if err != nil {
		fmt.Print("error retrieving Platform information")
		return errors.New("error Retrieving Platform information")
	}

	err = cacheClient.Add("access_token", token, 3600*time.Second)
	if err != nil {
		return errors.New("error Occured while caching token")
	}
	slackActionAttachmentOption := []slack.AttachmentActionOption{}

	for _, v := range platformDetails.User.ContributorOfOrganizations {
		cacheClient.Add(v.Name, v.Id, 3600*time.Second)
		slackActionAttachmentOption = append(slackActionAttachmentOption, slack.AttachmentActionOption{
			Text:  v.Name,
			Value: v.Id,
		})
	}

	slackAttachment := slack.Attachment{

		Pretext: "You are logged in to platform. Choose The Business Group ",
		Actions: []slack.AttachmentAction{
			{
				Options: slackActionAttachmentOption,
				Type:    slack.ActionType(slack.InputTypeSelect),
				Text:    "Choose The Business Group",
				Name:    "Business-Group",
			},
		},
	}

	slackClient.PostMessage(os.Getenv("CHANNEL_ID"), slack.MsgOptionAttachments(slackAttachment))
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
