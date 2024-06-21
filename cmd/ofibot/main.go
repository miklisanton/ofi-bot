package main

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"net/http"
	"ofibot/db/drivers"
	"ofibot/handler"
	"ofibot/repositories"
	"ofibot/service"
	"os"
)

var (
	DB     *sql.DB
	bot    *tgbotapi.BotAPI
	chatID int64
)

func init() {
	var err error
	chatID = 6220472117
	token := "6690181072:AAFLzMJ13agC-kVB9ZeOzyGw8CvqA_aHZNU"

	bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}

	connectionURL := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
		"antonmiklis",
		"1111",
		"localhost:5432",
		"ofi_bot",
		"disable")

	DB, err = postgres.Connect(connectionURL)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	tickerR := repositories.NewTickerRepo(DB)
	tickerS := service.NewTickerService(tickerR)
	userR := repositories.NewUserRepo(DB)
	userS := service.NewUserService(userR)
	userSubR := repositories.NewUserSubRepo(DB)
	userSubS := service.NewUserSubService(userSubR)
	subscriptionR := repositories.NewSubscriptionRepo(DB)
	subscriptionS := service.NewSubscriptionService(subscriptionR)

	botHandler := handler.NewBotHandler(tickerS, userS, userSubS, subscriptionS, bot)

	caBytes, err := os.ReadFile("ca.crt")
	if err != nil {
		log.Fatal(err)
	}

	ca := x509.NewCertPool()
	if !ca.AppendCertsFromPEM(caBytes) {
		log.Fatal("failed to append ca cert")
	}
	server := http.Server{
		Addr: ":443",
		TLSConfig: &tls.Config{
			ClientAuth: tls.RequireAndVerifyClientCert,
			ClientCAs:  ca,
			MinVersion: tls.VersionTLS13,
		},
	}

	http.HandleFunc("/webhook", botHandler.HandleWebhook)
	http.HandleFunc("/symbols", botHandler.HandleSymbols)

	port := "8080"

	log.Printf("Starting server on port %s\n", port)
	go func() {
		if err := server.ListenAndServeTLS("server.crt", "server.key"); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	//bot.Debug = true
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatalf("Failed to get updates: %v", err)
	}

	for update := range updates {
		botHandler.HandleUpdate(update)
	}

}
