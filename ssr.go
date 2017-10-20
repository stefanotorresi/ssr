package main

import (
	"encoding/json"
	"fmt"
	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

const slackApiBaseUrl = "https://api.slack.com/api/"
const channelMembersEndpoint = slackApiBaseUrl + "conversations.members"
const userInfoEndpoint = slackApiBaseUrl + "users.info"

type config struct {
	SlackAuthToken string `env:"SLACK_AUTH_TOKEN,required"`
	ChannelId      string `env:"CHANNEL_ID,required"`
}

type SlackResponse struct {
	Ok            bool   `json:"ok"`
	Error         string `json:"error"`
	RequestMethod string `json:"req_method"`
}

type ChannelMembers struct {
	SlackResponse
	Members []string `json:"members"`
}

type UserInfo struct {
	SlackResponse
	User struct {
		Id      string `json:"id"`
		Profile struct {
			DisplayName        string `json:"display_name"`
			RealName           string `json:"real_name"`
			RealNameNormalized string `json:"real_name_normalized"`
			Email              string `json:"email"`
		} `json:"profile"`
	} `json:"user"`
}

type ApiError struct {
	msg string
}

func (e *ApiError) Error() string { return "API Error: " + e.msg }

func init() {
	rand.Seed(time.Now().Unix())

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file: " + err.Error())
	}
}

func main() {

	config := config{}
	err := env.Parse(&config)
	exitOnError(err)

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	channelInfo, err := fetchChannelMembers(client, config)
	exitOnError(err)

	members := shuffle(channelInfo.Members)

	userInfos := make([]UserInfo, len(members))
	for i, memberId := range members {
		results := make(chan UserInfo, len(members))
		go func() {
			userInfo, err := fetchUserInfo(client, memberId, config.SlackAuthToken)
			exitOnError(err)
			results <- userInfo
		}()
		userInfos[i] = <-results
	}

	for _, user := range userInfos {
		fmt.Printf("%#v\n", user.User)
	}
}

func fetchChannelMembers(client *http.Client, config config) (ChannelMembers, error) {
	response, err := client.PostForm(channelMembersEndpoint, url.Values{
		"token":   {config.SlackAuthToken},
		"channel": {config.ChannelId},
	})
	if err != nil {
		return ChannelMembers{}, err
	}

	defer response.Body.Close()

	return decodeChannelMembers(response)
}

func fetchUserInfo(client *http.Client, memberId string, authToken string) (UserInfo, error) {
	response, err := client.PostForm(userInfoEndpoint, url.Values{
		"token": {authToken},
		"user":  {memberId},
	})
	if err != nil {
		return UserInfo{}, err
	}

	defer response.Body.Close()

	return decodeUserInfo(response)
}

func decodeChannelMembers(response *http.Response) (ChannelMembers, error) {
	channelMembers := ChannelMembers{}
	err := json.NewDecoder(response.Body).Decode(&channelMembers)

	if err == nil && !channelMembers.Ok {
		err = &ApiError{channelMembers.Error}
	}

	return channelMembers, err
}

func decodeUserInfo(response *http.Response) (UserInfo, error) {
	userInfo := UserInfo{}
	body, err := ioutil.ReadAll(response.Body)

	err = json.Unmarshal(body, &userInfo)

	if err == nil && !(userInfo.Ok && userInfo.User.Id != "") {
		err = &ApiError{fmt.Sprintf("%#v", userInfo)}
	}

	return userInfo, err
}

func exitOnError(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}

func shuffle(slice []string) []string {
	shuffled := make([]string, len(slice))
	copy(shuffled, slice)
	for i := range shuffled {
		j := rand.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}
	return shuffled
}
