package main

import (
	"fmt"
	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
	"github.com/stefanotorresi/ssr/slack"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type config struct {
	SlackAuthToken string `env:"SLACK_AUTH_TOKEN,required"`
	ChannelId      string `env:"CHANNEL_ID,required"`
}

func init() {
	rand.Seed(time.Now().Unix())

	err := godotenv.Load()
	if err != nil {
		log.Fatal("ApiError loading environment: " + err.Error())
	}
}

func main() {
	config := config{}
	err := env.Parse(&config)
	exitOnError(err)

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	api := slack.New(config.SlackAuthToken, client)

	channelInfo, err := api.FetchChannelMembers(config.ChannelId)
	exitOnError(err)

	members := shuffle(channelInfo.Members)
	fmt.Printf("%v\n", members)
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
