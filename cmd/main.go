package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"

	"go.uber.org/zap"

	"github.com/dafraer/openvpn-connections-bot/bot"
	"github.com/dafraer/openvpn-connections-bot/notifier"
	"github.com/dafraer/openvpn-connections-bot/tracker"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}
	token := os.Getenv("TOKEN")
	strUserID := os.Getenv("USER_ID")
	userID, err := strconv.ParseInt(strUserID, 10, 64)
	if err != nil {
		log.Fatalf("Error getting owner ID: %v", err)
	}
	if token == "" {
		panic("token is not specified")
	}

	//Declare context that is marked Done when os.Interrupt is called
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	//Create logger

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	sugar := logger.Sugar()

	//Create channels
	msg := make(chan notifier.Message)
	newAddr := make(chan string)
	reqAddr := make(chan string)
	respAddr := make(chan []string)
	t := tracker.New(newAddr, reqAddr, respAddr)

	//Create notifier
	n := notifier.New(userID, msg, sugar, t)

	//Create bot
	myBot, err := bot.New(token, sugar, n, msg)
	if err != nil {
		panic(err)
	}

	myBot.Run(ctx)
}
