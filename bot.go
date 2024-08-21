package telegrambot

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// Bot struct represents a Telegram Bot
type Bot struct {
	Token       string
	Client      *http.Client
	commands    map[string]func(chatID int64) string
	mu          sync.Mutex
}

// NewBot creates a new Telegram Bot instance
func NewBot(token string) *Bot {
	return &Bot{
		Token:    token,
		Client:   &http.Client{Timeout: 30 * time.Second},
		commands: make(map[string]func(chatID int64) string),
	}
}

// HandleUpdate processes incoming updates from Telegram
func (b *Bot) HandleUpdate(update Update) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if handler, exists := b.commands[update.Message.Text]; exists {
		response := handler(update.Message.Chat.ID)
		b.SendMessage(update.Message.Chat.ID, response)
	}

	return nil
}

// AddCommand allows users to add custom commands and their handlers
func (b *Bot) AddCommand(command string, handler func(chatID int64) string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.commands[command] = handler
}

// SendMessage sends a message to a specific chat ID
func (b *Bot) SendMessage(chatID int64, text string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", b.Token)
	data := url.Values{}
	data.Set("chat_id", fmt.Sprintf("%d", chatID))
	data.Set("text", text)

	resp, err := b.Client.PostForm(apiURL, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to send message")
	}

	return nil
}

// GetUpdates retrieves updates from the bot
func (b *Bot) GetUpdates(offset int) ([]Update, error) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates", b.Token)
	data := url.Values{}
	data.Set("offset", fmt.Sprintf("%d", offset))

	resp, err := b.Client.PostForm(apiURL, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response UpdateResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	if !response.Ok {
		return nil, errors.New("failed to get updates")
	}

	return response.Result, nil
}
