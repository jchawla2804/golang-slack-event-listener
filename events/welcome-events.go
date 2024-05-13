package events

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
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

func HandleSlackAppMentions(appMentionEvent *slackevents.AppMentionEvent, slackClient *slack.Client) error {
	slackUser, err := slackClient.GetUserInfo(appMentionEvent.User)
	if err != nil {

		log.Printf("Err 1. Line 179 : %s", err.Error())
		return err
	}

	text := strings.ToLower(appMentionEvent.Text)
	slackAttachment := slack.Attachment{}

	buttonBlockElement1 := slack.NewButtonBlockElement("button1", "basic-auth", &slack.TextBlockObject{Type: slack.PlainTextType, Text: "Login using Basic Auth"})
	buttonBlockElement2 := slack.NewButtonBlockElement("button2", "connected-app", &slack.TextBlockObject{Type: slack.PlainTextType, Text: "Login using Connected App"})

	blockSet := []slack.Block{
		//sectionObject,
		slack.NewSectionBlock(&slack.TextBlockObject{Type: slack.MarkdownType, Text: "Please Choose login option"}, nil, nil),
		slack.NewActionBlock("actionblock789", buttonBlockElement1, buttonBlockElement2),
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

	_, timestamp, err := slackClient.PostMessage(os.Getenv("CHANNEL_ID"), slack.MsgOptionBlocks(blockSet...))

	if err != nil {
		log.Println("Error Happened While sending message")
		log.Print(err.Error())
		return err
	}
	fmt.Printf("Message Was Sent at this time %s", timestamp)
	return nil

}
