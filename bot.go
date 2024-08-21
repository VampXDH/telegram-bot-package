package telegrambot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Bot represents the Telegram bot
type Bot struct {
	Token  string
	BaseURL string
}

// NewBot creates a new Telegram bot
func NewBot(token string) *Bot {
	return &Bot{
		Token:  token,
		BaseURL: fmt.Sprintf("https://api.telegram.org/bot%s/", token),
	}
}

// SendMessage sends a message to a specific chat ID
func (b *Bot) SendMessage(chatID int64, text string) error {
	url := b.BaseURL + "sendMessage"
	message := map[string]interface{}{
		"chat_id": chatID,
		"text":    text,
	}
	messageJSON, err := json.Marshal(message)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(messageJSON))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send message: %s", resp.Status)
	}

	return nil
}
