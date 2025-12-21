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

const (
	logLevelDebug = "debug"
	logLevelInfo  = "info"
	logLevelWarn  = "warn"
	logLevelError = "error"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}
	token := os.Getenv("TOKEN")
	strUserID := os.Getenv("USER_ID")
	userID, err := strconv.ParseInt(strUserID, 10, 64)
	logLevel := os.Getenv("info")
	if err != nil {
		log.Fatalf("Error getting owner ID: %v", err)
	}
	if token == "" {
		panic("token is not specified")
	}
	if logLevel == "" {
		logLevel = "info"
	}

	//Declare context that is marked Done when os.Interrupt is called
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	//Create logger
	loggerCfg := zap.NewDevelopmentConfig()
	switch logLevel {
	case logLevelDebug:
		loggerCfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case logLevelInfo:
		loggerCfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case logLevelWarn:
		loggerCfg.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case logLevelError:
		loggerCfg.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		loggerCfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}
	logger, err := loggerCfg.Build()
	if err != nil {
		panic(err)
	}
	sugar := logger.Sugar()

	//Create channels
	msg := make(chan notifier.Message)
	newAddr := make(chan string)
	reqAddr := make(chan string)
	respAddr := make(chan []string)
	t := tracker.New(newAddr, reqAddr, respAddr, sugar)

	//Create notifier
	n := notifier.New(userID, msg, sugar, t)

	//Create bot
	myBot, err := bot.New(token, sugar, n, msg)
	if err != nil {
		panic(err)
	}

	myBot.Run(ctx)
}
