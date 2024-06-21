package handler

import (
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"io"
	"log"
	"net/http"
	"ofibot/models"
	"ofibot/service"
	"strconv"
	"strings"
)

const tickersPerPage = 36

type BotHandler struct {
	tickerService       *service.TickerService
	userService         *service.UserService
	userSubService      *service.UserSubService
	subscriptionService *service.SubscriptionService
	bot                 *tgbotapi.BotAPI
}

func NewBotHandler(tickerService *service.TickerService,
	userService *service.UserService,
	userSubService *service.UserSubService,
	subscriptionService *service.SubscriptionService,
	bot *tgbotapi.BotAPI) *BotHandler {
	return &BotHandler{
		tickerService:       tickerService,
		userService:         userService,
		userSubService:      userSubService,
		subscriptionService: subscriptionService,
		bot:                 bot,
	}
}

// WebhookPayload represents the structure of the incoming webhook payload
type WebhookPayload struct {
	Message string  `json:"message"`
	OFI     float64 `json:"ofi"`
	Time    string  `json:"time"`
	Ticker  string  `json:"ticker"`
}

// HandleWebhook processes incoming webhook requests
func (h *BotHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
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

	ids, err := h.subscriptionService.GetUsersSubscribed(payload.Ticker)
	h.sendMultiple(ids, fmt.Sprintf("%s: %.2f", payload.Ticker, payload.OFI))
}

func (h *BotHandler) HandleSymbols(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	symbols, err := h.tickerService.GetSubscribedTickers()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	jsonData, err := json.Marshal(symbols)
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return
	}

	if _, err := w.Write(jsonData); err != nil {
		log.Println("Error writing response:", err)
		return
	}
}

func (h *BotHandler) HandleUpdate(update tgbotapi.Update) {
	if update.Message != nil {
		chatID := update.Message.Chat.ID
		switch update.Message.Text {
		case "/start":
			if err := h.userService.AddUser(chatID); err != nil {
				log.Println("Failed to add user: ", err)
				return
			}
			h.sendWelcomeMessage(chatID)
		case "Unsubscribe":
			tickers, err := h.subscriptionService.GetUserTickers(chatID)
			if err != nil {
				log.Println("Failed to get tickers: ", err)
				return
			}
			h.sendTickers(chatID, 0, tickers, true)
		case "Choose Tickers":
			tickers, err := h.tickerService.GetAllTickers()
			if err != nil {
				log.Println("Failed to get tickers: ", err)
				return
			}
			h.sendTickers(chatID, 0, tickers, false)
		case "Subscribed tickers":
			tickers, err := h.subscriptionService.GetUserTickers(chatID)
			if err != nil {
				log.Println(err)
				return
			}
			if len(tickers) == 0 {
				h.sendMessage(chatID, "No subscribed tickers")
				h.sendWelcomeMessage(chatID)
				return
			}

			msgString, err := h.subscriptionService.GetSubscriptionsString(tickers)
			if err != nil {
				log.Println(err)
				return
			}

			msg := tgbotapi.NewMessage(chatID, msgString)
			keyboard := tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("Unsubscribe"),
				),
			)
			msg.ReplyMarkup = keyboard
			if _, err := h.bot.Send(msg); err != nil {
				log.Printf("Failed to send unsubscribe message: %v", err)
			}
		default:
			h.sendMessage(chatID, "Unknown command")
		}
	} else if update.CallbackQuery != nil {
		h.handleCallbackQuery(update.CallbackQuery)
	}
}

func (h *BotHandler) sendWelcomeMessage(chatID int64) {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Choose Tickers"),
			tgbotapi.NewKeyboardButton("Subscribed tickers"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, "Choose an option:")
	msg.ReplyMarkup = keyboard

	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("Failed to send welcome message: %v", err)
	}
}

func (h *BotHandler) handleCallbackQuery(callbackQuery *tgbotapi.CallbackQuery) {
	data := callbackQuery.Data
	chatID := callbackQuery.Message.Chat.ID
	if strings.HasPrefix(data, "-") {
		err := h.subscriptionService.RemoveSubscription(chatID, data[1:])
		if err != nil {
			log.Println("Failed to remove subscription: ", err)
			return
		}

		h.sendMessage(chatID, "Subscription removed: "+data[1:])
		h.sendWelcomeMessage(chatID)
		return
	}
	if strings.HasPrefix(data, "page_") {
		page, err := strconv.Atoi(strings.TrimPrefix(data, "page_"))
		if err != nil {
			log.Printf("Failed to parse page number: %v", err)
			return
		}

		tickers, err := h.tickerService.GetAllTickers()
		if err != nil {
			log.Println(err)
			return
		}
		h.sendTickers(chatID, page, tickers, false)
	} else {
		sub, err := h.subscriptionService.AddSubscription(chatID, data)
		if err != nil {
			log.Printf("Failed to add subscription: %v", err)
			return
		}

		var msgContent string

		if sub == nil {
			msgContent = "You have already subscribed to " + data
		} else {
			msgContent = "You have subscribed to " + sub.TickerID
		}
		msg := tgbotapi.NewMessage(chatID, msgContent)
		if _, err := h.bot.Send(msg); err != nil {
			log.Printf("Failed to send callback response: %v", err)
		}
	}
}

func (h *BotHandler) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("Failed to send message: %v", err)
	}
}

func (h *BotHandler) sendMultiple(ids []int64, text string) {
	for _, id := range ids {
		h.sendMessage(id, text)
	}
}

func (h *BotHandler) sendTickers(chatID int64, page int, tickers []models.Ticker, unsub bool) {

	start := page * tickersPerPage
	end := start + tickersPerPage
	if end > len(tickers) {
		end = len(tickers)
	}

	var rows [][]tgbotapi.InlineKeyboardButton

	for i := start; i < end; i += 3 {
		row := []tgbotapi.InlineKeyboardButton{}
		for j := 0; j < 3 && i+j < end; j++ {
			ticker := tickers[i+j]

			// Add - sign for unsubscribing
			var data string
			if unsub == true {
				data = "-" + ticker.Symbol
			} else {
				data = ticker.Symbol
			}
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(ticker.Symbol, data))
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(row...))
	}

	// Add navigation buttons if necessary
	var navButtons []tgbotapi.InlineKeyboardButton
	if page > 0 {
		navButtons = append(navButtons, tgbotapi.NewInlineKeyboardButtonData("Previous", fmt.Sprintf("page_%d", page-1)))
	}
	if end < len(tickers) {
		navButtons = append(navButtons, tgbotapi.NewInlineKeyboardButtonData("Next", fmt.Sprintf("page_%d", page+1)))
	}
	if len(navButtons) > 0 {
		rows = append(rows, navButtons)
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	msg := tgbotapi.NewMessage(chatID, "Please select tickers:")
	msg.ReplyMarkup = keyboard

	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("Failed to send tickers: %v", err)
	}
}
