package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	apps "github.com/mattermost/mattermost-plugin-apps/upstream/upaws"
	mmmodel "github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
)

func send(webhookURL string, payload mmmodel.CommandResponse) error {
	marshalContent, _ := json.Marshal(payload)
	var jsonStr = []byte(marshalContent)
	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonStr))
	req.Header.Set("X-Custom-Header", "aws-sns")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func sendAppDeploymentNotification(deployData *apps.DeployData, bundle string) error {
	var fields []*mmmodel.SlackAttachmentField

	fields = append(fields, &mmmodel.SlackAttachmentField{
		Title: "Name",
		Value: fmt.Sprintf("[%s](%s)", deployData.Manifest.DisplayName, deployData.Manifest.HomepageURL),
		Short: true,
	})
	fields = append(fields, &mmmodel.SlackAttachmentField{
		Title: "ID",
		Value: fmt.Sprintf("`%s`", deployData.Manifest.AppID),
		Short: true,
	})
	fields = append(fields, &mmmodel.SlackAttachmentField{
		Title: "Version",
		Value: deployData.Manifest.Version,
		Short: true,
	})
	fields = append(fields, &mmmodel.SlackAttachmentField{
		Title: "Bundle",
		Value: fmt.Sprintf("`%s`", bundle),
		Short: true,
	})

	fields = append(fields, &mmmodel.SlackAttachmentField{Title: "Environment", Value: os.Getenv("Environment"), Short: false})

	attachment := &mmmodel.SlackAttachment{
		Color:  "#006400",
		Fields: fields,
		Title:  "A Mattermost apps was successfully deployed/updated",
	}

	payload := mmmodel.CommandResponse{
		Username:    "Mattermost Apps Deployer",
		IconURL:     "https://cdn-images-1.medium.com/max/1200/1*9860tn6_CPEPnBxF1wIpmw@2x.jpeg",
		Attachments: []*mmmodel.SlackAttachment{attachment},
	}
	err := send(os.Getenv("MattermostNotificationsHook"), payload)
	if err != nil {
		return errors.Wrap(err, "failed tο send Mattermost request payload")
	}
	return nil
}

func sendMattermostErrorNotification(errorMessage error, message string) error {
	attachment := &mmmodel.SlackAttachment{
		Color: "#FF0000",
		Fields: []*mmmodel.SlackAttachmentField{
			{Title: message, Short: false},
			{Title: "Error Message", Value: errorMessage.Error(), Short: false},
			{Title: "Environment", Value: os.Getenv("Environment"), Short: true},
		},
	}

	payload := mmmodel.CommandResponse{
		Username:    "Mattermost Apps Deployer",
		IconURL:     "https://cdn-images-1.medium.com/max/1200/1*9860tn6_CPEPnBxF1wIpmw@2x.jpeg",
		Attachments: []*mmmodel.SlackAttachment{attachment},
	}
	err := send(os.Getenv("MattermostAlertsHook"), payload)
	if err != nil {
		return errors.Wrap(err, "failed tο send Mattermost error payload")
	}
	return nil
}
