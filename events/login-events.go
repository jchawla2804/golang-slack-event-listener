package events

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jchawla2804/golang-slack-event-listener/helper"
	"github.com/slack-go/slack"
)

func HandleInteractiveDialogBoxEvent(slackClient *slack.Client, buttonValue, triggerId string) error {
	modal := slack.ModalViewRequest{}
	modal.Title = &slack.TextBlockObject{Type: slack.PlainTextType, Text: "Login To Platform"}
	modal.Submit = &slack.TextBlockObject{Type: slack.PlainTextType, Text: "Submit"}
	modal.Type = slack.VTModal
	var blockSet []slack.Block
	log.Println("Button Value ", buttonValue)
	if buttonValue == "basic-auth" {
		blockSet = []slack.Block{
			slack.NewInputBlock("username", &slack.TextBlockObject{Type: slack.PlainTextType, Text: "Enter Username"}, nil, slack.PlainTextInputBlockElement{Type: slack.METPlainTextInput, ActionID: "user"}),
			slack.NewInputBlock("password", &slack.TextBlockObject{Type: slack.PlainTextType, Text: "Enter Password"}, nil, slack.PlainTextInputBlockElement{Type: slack.METPlainTextInput, ActionID: "pass"}),
		}
	} else {
		blockSet = []slack.Block{
			slack.NewInputBlock("clientId", &slack.TextBlockObject{Type: slack.PlainTextType, Text: "Enter Connected app client Id"}, nil, slack.PlainTextInputBlockElement{Type: slack.METPlainTextInput, ActionID: "clientId"}),
			slack.NewInputBlock("clientSecret", &slack.TextBlockObject{Type: slack.PlainTextType, Text: "Enter Connected app client secret"}, nil, slack.PlainTextInputBlockElement{Type: slack.METPlainTextInput, ActionID: "clientSecret"}),
		}
	}
	modal.Blocks = slack.Blocks{BlockSet: blockSet}

	//_, timestamp, err := slackClient.PostMessage(os.Getenv("CHANNEL_ID"), slack)

	_, err := slackClient.OpenView(triggerId, modal)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	return nil

}

func HandleLogin(slackClient *slack.Client, username, password, typeOfAuth string) error {

	log.Println("Type of Auth ", typeOfAuth)
	token, err := helper.GetToken(username, password, typeOfAuth)
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

	var businessGroupOptions []*slack.OptionBlockObject

	for _, v := range platformDetails.User.ContributorOfOrganizations {
		businessGroupOptions = append(businessGroupOptions, slack.NewOptionBlockObject(v.Id, slack.NewTextBlockObject("plain_text", v.Name, false, false), nil))

	}

	slackSelectBlockElement := slack.NewOptionsSelectBlockElement(slack.OptTypeStatic, slack.NewTextBlockObject("plain_text", "Choose Business Group", false, false), "select2", businessGroupOptions...)

	block := slack.NewActionBlock("bg-block", slackSelectBlockElement)

	sectionBlock := slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", "You are logged in to platform. Choose The Business Group", false, false), nil, nil)

	_, _, err = slackClient.PostMessage(os.Getenv("CHANNEL_ID"), slack.MsgOptionBlocks(sectionBlock, block))
	if err != nil {
		return err
	}
	return nil

}
