package main

import (
	"log"
	"net/http"
	"time"

	"github.com/spf13/viper"
	"github.com/turnage/graw/reddit"
)

func main() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Panic("Could not read configuration file")
	}

	// finnhubApiKey := viper.GetString("finnhub.api.key")
	// if finnhubApiKey == "" {
	// 	log.Panic("Could not get API Key from configuration file")
	// }

	app := reddit.App{
		ID:     viper.GetString("reddit.api.app-id"),
		Secret: viper.GetString("reddit.api.app-secret"),
	}

	botConfig := reddit.BotConfig{
		Agent:  "golang-Script",
		App:    app,
		Rate:   time.Second,
		Client: &http.Client{},
	}

	myBot, err := reddit.NewBot(botConfig)
	if err != nil {
		log.Panic("Could not set reddit bot")
	}

	harvest, err := myBot.Listing("/r/wallstreetbets/comments/ly9gxp", "")
	if err != nil {
		log.Panic("Could not get harvest")
	}
	_ = harvest
	// harvest, err := myBot.Listing("/r/wallstreetbets", "")
	// if err != nil {
	// 	log.Panic("Could not get harvest")
	// }

	// for _, post := range harvest.Posts {
	// 	log.Println(post.Name)
	// 	log.Println(post.Title)
	// 	log.Println(time.Unix(int64(post.CreatedUTC), 0).Format(time.RFC3339))
	// 	log.Println("Comments:")
	// 	log.Print(post.NumComments)

	// 	for _, comment := range post.Replies {
	// 		log.Print(comment.Body)
	// 		log.Println(time.Unix(int64(comment.CreatedUTC), 0).Format(time.RFC3339))
	// 	}

	// 	log.Println("--------------------------------")
	// }
}
