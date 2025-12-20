package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"

	"go.uber.org/zap"

	"github.com/dafraer/openvpn-connections-bot/bot"
	"github.com/joho/godotenv"
)

const defaultLogPath = "./log/logs.txt"

func main() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}
	token := os.Getenv("TOKEN")
	strUserID := os.Getenv("USER_ID")
	userID, err := strconv.Atoi(strUserID)
	if err != nil {
		log.Fatalf("Error getting owner ID: %v", err)
	}
	logPath := os.Getenv("LOG_PATH")
	if token == "" {
		panic("token is not specified")
	}
	if logPath == "" {
		logPath = defaultLogPath
	}

	//Declare context that is marked Done when os.Interrupt is called
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	//Create logger
	cfg := zap.NewDevelopmentConfig()
	cfg.Development = true
	cfg.OutputPaths = []string{logPath}
	cfg.ErrorOutputPaths = []string{logPath}

	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	sugar := logger.Sugar()

	//Create bot
	myBot, err := bot.New(token, sugar, int64(userID))
	if err != nil {
		panic(err)
	}

	myBot.Run(ctx)
}
