package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var bot *tgbotapi.BotAPI
var chatID int64

// WebhookPayload represents the structure of the incoming webhook payload
type WebhookPayload struct {
	Message string  `json:"message"`
	OFI     float64 `json:"ofi"`
	Time    string  `json:"time"`
}

// handleWebhook processes incoming webhook requests
func handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var payload WebhookPayload
	err = json.Unmarshal(body, &payload)
	if err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Received webhook: %+v\n", payload)

	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("%.2f", payload.OFI))
	_, err = bot.Send(msg)
	if err != nil {
		log.Printf("Failed to send message: %+v\n", err)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome!\n")
}

func init() {
	var err error
	chatID = 6220472117
	token := "6690181072:AAFLzMJ13agC-kVB9ZeOzyGw8CvqA_aHZNU"
	bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}
}
func main() {
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
	http.HandleFunc("/webhook", handleWebhook)
	http.HandleFunc("/", index)

	bot.Debug = true

	port := "8080"
	log.Printf("Starting server on port %s\n", port)
	if err := server.ListenAndServeTLS("server.crt", "server.key"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
