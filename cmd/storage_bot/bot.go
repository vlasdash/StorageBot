package main

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/vlasdash/passwordbot/config"
	"github.com/vlasdash/passwordbot/init/db"
	"github.com/vlasdash/passwordbot/internal/credential"
	"github.com/vlasdash/passwordbot/pkg/handler"
	"log"
	"net/http"
	"os"
)

const ConfigPath = "./config/"

func startPasswordStorageBot(ctx context.Context) error {
	_ = godotenv.Load(".env.example")

	err := config.LoadConfig(ConfigPath)
	if err != nil {
		log.Printf("read config failed: %v\n", err)
		return err
	}

	botToken := os.Getenv("BOT_TOKEN")
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Printf("create BotAPI instance failed: %v\n", err)
		return err
	}

	wh, err := tgbotapi.NewWebhook(config.C.App.WebhookURL)
	if err != nil {
		log.Printf("create new webhook failed: %v\n", err)
		return err
	}

	_, err = bot.Request(wh)
	if err != nil {
		log.Printf("set webhook failed: %v\n", err)
		return err
	}

	updates := bot.ListenForWebhook("/")

	go func() {
		log.Printf(
			"http err: %v\n",
			http.ListenAndServe(fmt.Sprintf(":%d", config.C.App.Port), nil))
	}()

	mysqlDB, err := db.InitMySQL()
	if err != nil {
		return err
	}
	defer func() {
		if err = mysqlDB.Close(); err != nil {
			log.Printf("error at close mysql connection: %v\n", err)
		}
	}()
	credentialRepo := credential.NewMySQLRepo(mysqlDB)
	secretKey := os.Getenv("SECRET_KEY")
	coder := credential.NewCryptCoder(secretKey)
	credentialHandler := handler.NewHandler(bot, credentialRepo, coder)

	for {
		select {
		case <-ctx.Done():
			return nil
		case update := <-updates:
			credentialHandler.HandleCommand(update)
		}
	}
}

func main() {
	err := startPasswordStorageBot(context.Background())
	if err != nil {
		panic(err)
	}
}
