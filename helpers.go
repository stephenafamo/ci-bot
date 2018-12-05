package main

import (
	"encoding/json"
)

func sendSuccessProdDeploy(payload actionPayload, user, url string) (err error) {

	project := payload.Build.Project

	var newM SlackMessage
	newM.Channel = project.Channel
	newM.Text = "New production deployment for project " + project.Name + " by <@" + user + ">"

	newM.Attachments = []SlackAttachment{
		SlackAttachment{
			Fallback: "Project: " + project.Name + " Image: " + payload.Build.Image + " By: <@" + user + ">",
			Fields: []SlackField{
				SlackField{
					Title: "Project",
					Value: project.Name,
					Short: false,
				},
				SlackField{
					Title: "Docker Image",
					Value: payload.Build.Image,
					Short: false,
				},
				SlackField{
					Title: "By",
					Value: "<@" + user + ">",
					Short: true,
				},
			},
		},
		SlackAttachment{
			Fallback: "View project: " + url,
			Color:    "good",
			Actions: []SlackAction{
				SlackAction{
					Type: "button",
					Text: "View project",
					URL:  url,
				},
			},
		},
	}

	_, err = sendSlack(newM)
	return
}

func sendFailedProdDeploy(payload actionPayload, deployErr error) (err error) {

	project := payload.Build.Project

	var newM SlackMessage
	newM.Channel = project.Channel
	newM.Text = "Production deployment failed for project " + project.Name

	newM.Attachments = []SlackAttachment{
		SlackAttachment{
			Fallback: "Project: " + project.Name + " Image: " + payload.Build.Image,
			Fields: []SlackField{
				SlackField{
					Title: "Project",
					Value: project.Name,
					Short: false,
				},
				SlackField{
					Title: "Docker Image",
					Value: payload.Build.Image,
					Short: false,
				},
			},
		},
		SlackAttachment{
			Fallback: "Failure Reason: " + deployErr.Error(),
			Color:    "danger",
			Fields: []SlackField{
				SlackField{
					Title: "Failure Reason",
					Value: deployErr.Error(),
					Short: false,
				},
			},
		},
	}

	_, err = sendSlack(newM)
	return
}

func sendAttemptDeployMessage(build Build) (ts string, err error) {
	msg := getAttemptDeployMessage(build)

	resp, err := sendSlack(msg)
	if err != nil {
		return
	}

	attemptMsgResp := make(map[string]string)
	json.Unmarshal(resp, &attemptMsgResp)
	ts = attemptMsgResp["ts"]

	return
}

func getAttemptDeployMessage(build Build) SlackMessage {
	return SlackMessage{
		Channel: build.Project.Channel,
		Text:    "New Build complete.\nAttempting deployment...",
		Attachments: []SlackAttachment{
			SlackAttachment{
				Fallback: "Project: " + build.Project.Name + " Type: " + build.Type + " Target: " + build.Target + " Image: " + build.Image,
				Fields: []SlackField{
					SlackField{
						Title: "Project",
						Value: build.Project.Name,
						Short: false,
					},
					SlackField{
						Title: "Docker Image",
						Value: build.Image,
						Short: false,
					},
					SlackField{
						Title: "Type",
						Value: build.Type,
						Short: true,
					},
					SlackField{
						Title: "Target",
						Value: build.Target,
						Short: true,
					},
				},
			},
		},
	}
}

func sendDeploySuccessMessage(build Build, ts, url string) (err error) {
	msg := getDeploySuccessMessage(build, url)
	msg.Update = true
	msg.Ts = ts

	_, err = sendSlack(msg)
	return
}

func getDeploySuccessMessage(build Build, url string) SlackMessage {
	return SlackMessage{
		Channel: build.Project.Channel,
		Text:    "New Build complete.\nDeployment Successful! :sunglasses:",
		Attachments: []SlackAttachment{
			SlackAttachment{
				Fallback: "Project: " + build.Project.Name + " Type: " + build.Type + " Target: " + build.Target + " Image: " + build.Image,
				Fields: []SlackField{
					SlackField{
						Title: "Project",
						Value: build.Project.Name,
						Short: false,
					},
					SlackField{
						Title: "Docker Image",
						Value: build.Image,
						Short: false,
					},
					SlackField{
						Title: "Type",
						Value: build.Type,
						Short: true,
					},
					SlackField{
						Title: "Target",
						Value: build.Target,
						Short: true,
					},
				},
			},
			SlackAttachment{
				Fallback: "View project: " + url,
				Color:    "good",
				Actions: []SlackAction{
					SlackAction{
						Type: "button",
						Text: "View project",
						URL:  url,
					},
				},
			},
		},
	}
}

func sendFailedDeployMessage(build Build, ts string, deployErr error) (err error) {
	msg := getFailedDeployMessage(build, deployErr)
	msg.Update = true
	msg.Ts = ts

	_, err = sendSlack(msg)
	return
}

func getFailedDeployMessage(build Build, err error) SlackMessage {
	return SlackMessage{
		Channel: build.Project.Channel,
		Text:    "New Build complete.\nDeployment Failed :sob:",
		Attachments: []SlackAttachment{
			SlackAttachment{
				Fallback: "Project: " + build.Project.Name + " Type: " + build.Type + " Target: " + build.Target + " Image: " + build.Image,
				Fields: []SlackField{
					SlackField{
						Title: "Project",
						Value: build.Project.Name,
						Short: false,
					},
					SlackField{
						Title: "Docker Image",
						Value: build.Image,
						Short: false,
					},
					SlackField{
						Title: "Type",
						Value: build.Type,
						Short: true,
					},
					SlackField{
						Title: "Target",
						Value: build.Target,
						Short: true,
					},
				},
			},
			SlackAttachment{
				Fallback: "Failure Reason: " + err.Error(),
				Color:    "danger",
				Fields: []SlackField{
					SlackField{
						Title: "Failure Reason",
						Value: err.Error(),
						Short: false,
					},
				},
			},
		},
	}
}

func sendOwnerMessages(build Build, url string) (
	payload actionPayload, errs []error) {
	var oMsgs []ownerMsg

	successMessage := getDeploySuccessMessage(build, url)

	for _, user := range build.Project.Owners {
		successMessage.Channel = user
		resp, err := sendSlack(successMessage)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		respMap := make(map[string]string)
		json.Unmarshal(resp, &respMap)

		oMsgs = append(oMsgs, ownerMsg{
			Owner:   user,
			Ts:      respMap["ts"],
			Channel: respMap["channel"],
		})
	}

	payload = actionPayload{
		Build:         build,
		OwnerMessages: oMsgs,
	}

	OwnerMessage, err := getOwnerMessage(build, url, payload)
	if err != nil {
		errs = append(errs, err)
		return
	}

	for _, oM := range payload.OwnerMessages {
		OwnerMessage.Update = true
		OwnerMessage.Channel = oM.Channel
		OwnerMessage.Ts = oM.Ts

		_, err := sendSlack(OwnerMessage)
		if err != nil {
			errs = append(errs, err)
			continue
		}
	}

	return
}

func getOwnerMessage(build Build, url string, payload actionPayload) (SlackMessage, error) {

	marshaledPayload, err := json.Marshal(payload)
	if err != nil {
		return SlackMessage{}, err
	}

	qaTeamAttachment := getQaSlackAttachment(build)

	message := SlackMessage{
		Text: "New Build complete.\nDeployment Successful! :sunglasses:",
		Attachments: []SlackAttachment{
			SlackAttachment{
				Fallback: "Project: " + build.Project.Name + " Type: " + build.Type + " Target: " + build.Target + " Image: " + build.Image,
				Fields: []SlackField{
					SlackField{
						Title: "Project",
						Value: build.Project.Name,
						Short: false,
					},
					SlackField{
						Title: "Docker Image",
						Value: build.Image,
						Short: false,
					},
					SlackField{
						Title: "Type",
						Value: build.Type,
						Short: true,
					},
					SlackField{
						Title: "Target",
						Value: build.Target,
						Short: true,
					},
				},
			},
			SlackAttachment{
				Fallback: "View project: " + url,
				Color:    "good",
				Actions: []SlackAction{
					SlackAction{
						Type: "button",
						Text: "View project",
						URL:  url,
					},
				},
			},
			qaTeamAttachment,
			SlackAttachment{
				Fallback:   "Deploy to Production.",
				CallbackID: "Deploy Decision",
				Actions: []SlackAction{
					SlackAction{
						Type:  "button",
						Text:  "Deploy to Production",
						Name:  "deploy",
						Value: string(marshaledPayload),
						Style: "primary",
						Confirm: map[string]string{
							"title":        "Are you sure?",
							"text":         "This will deploy to production. The process cannot be reversed.",
							"ok_text":      "Deploy",
							"dismiss_text": "Cancel",
						},
					},
					SlackAction{
						Type:  "button",
						Text:  "Close",
						Name:  "close",
						Value: string(marshaledPayload),
						Style: "danger",
						Confirm: map[string]string{
							"title":        "Are you sure?",
							"text":         "This will close this deployment. The process cannot be reversed.",
							"ok_text":      "Close",
							"dismiss_text": "Cancel",
						},
					},
				},
			},
		},
	}

	return message, nil
}

func sendQaMessages(build Build, url string, payload actionPayload) (errs []error) {

	QAmsg, err := getQAMessage(build, url, payload)
	if err != nil {
		errs = append(errs, err)
		return
	}

	for _, user := range build.Project.QA {
		QAmsg.Channel = user
		_, err := sendSlack(QAmsg)
		if err != nil {
			errs = append(errs, err)
			continue
		}
	}

	return
}

func getQAMessage(build Build, url string, payload actionPayload) (SlackMessage, error) {

	marshaledPayload, err := json.Marshal(payload)
	if err != nil {
		return SlackMessage{}, err
	}

	message := SlackMessage{
		Text: "New Build complete.\nDeployment Successful! :sunglasses:",
		Attachments: []SlackAttachment{
			SlackAttachment{
				Fallback: "Project: " + build.Project.Name + " Type: " + build.Type + " Target: " + build.Target + " Image: " + build.Image,
				Fields: []SlackField{
					SlackField{
						Title: "Project",
						Value: build.Project.Name,
						Short: false,
					},
					SlackField{
						Title: "Docker Image",
						Value: build.Image,
						Short: false,
					},
					SlackField{
						Title: "Type",
						Value: build.Type,
						Short: true,
					},
					SlackField{
						Title: "Target",
						Value: build.Target,
						Short: true,
					},
				},
			},
			SlackAttachment{
				Fallback: "View project: " + url,
				Color:    "good",
				Actions: []SlackAction{
					SlackAction{
						Type: "button",
						Text: "View project",
						URL:  url,
					},
				},
			},
			SlackAttachment{
				Title:      "Kindly perform QA for this project.",
				Fallback:   "Kindly perform QA for this project.",
				CallbackID: "QA Response",
				Actions: []SlackAction{
					SlackAction{
						Type:  "button",
						Text:  "Approve",
						Name:  "approve",
						Value: string(marshaledPayload),
						Style: "primary",
					},
					SlackAction{
						Type:  "button",
						Text:  "Reject",
						Name:  "reject",
						Value: string(marshaledPayload),
						Style: "danger",
					},
				},
			},
		},
	}

	return message, nil
}

func getQaSlackAttachment(build Build) SlackAttachment {

	qaTeamAttachment := SlackAttachment{
		Title:    "QA to be done by:",
		Fallback: "QA to be done by:",
	}

	for _, user := range build.Project.QA {
		qaTeamAttachment.Fields = append(qaTeamAttachment.Fields, SlackField{
			Value: "<@" + user + ">",
			Short: true,
		})
	}

	return qaTeamAttachment
}
