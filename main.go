package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kr/pretty"
	"github.com/nlopes/slack"
	"github.com/peterbourgon/diskv"
	"github.com/sirupsen/logrus"
)

const SlackTokenEnvKey = "SLACK_TOKEN"

func displayDiff(slackUser slack.User, diffs []string) {
	fmt.Printf("User %s has some update\n", slackUser.Name)
	for _, diff := range diffs {
		fmt.Println(diff)
	}
	fmt.Println("")
}

func displayNewUser(slackUser slack.User) {
	fmt.Printf("User %s was created\n", slackUser.Name)
	fmt.Println("")
}

func compareUser(slackUser, oldSlackUser slack.User) []string {
	// Remove call status
	if slackUser.Profile.StatusEmoji == ":slack_call:" {
		slackUser.Profile.StatusEmoji = oldSlackUser.Profile.StatusEmoji
		slackUser.Profile.StatusText = oldSlackUser.Profile.StatusText
	}
	slackUser.TZLabel = oldSlackUser.TZLabel
	slackUser.TZOffset = oldSlackUser.TZOffset
	slackUser.TZ = oldSlackUser.TZ
	slackUser.Updated = oldSlackUser.Updated
	return pretty.Diff(oldSlackUser, slackUser)
}

func main() {
	api := slack.New(os.Getenv(SlackTokenEnvKey))
	cache := diskv.New(diskv.Options{
		BasePath: "./user_cache",
	})

	users, err := api.GetUsers()
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	for _, slackUser := range users {
		jsonUser, err := json.Marshal(slackUser)
		if err != nil {
			logrus.Errorf("error when marshaling user: %v", slackUser)
		}
		if cache.Has(slackUser.ID) {
			oldSlackUserByte, err := cache.Read(slackUser.ID)
			if err != nil {
				logrus.Errorf("error when reading user from cache: %s", slackUser.ID)
			}
			oldSlackUser := slack.User{}
			err = json.Unmarshal(oldSlackUserByte, &oldSlackUser)
			if err != nil {
				logrus.Errorf("error when unmarshaling user from cache: %s", slackUser.ID)
			}

			diffs := compareUser(slackUser, oldSlackUser)
			if len(diffs) == 0 {
				continue
			}
			displayDiff(slackUser, diffs)

		} else {
			displayNewUser(slackUser)
		}
		cache.Write(slackUser.ID, jsonUser)

	}
}
