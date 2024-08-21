package telegrambot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Bot represents the Telegram bot
type Bot struct {
	Token   string
	BaseURL string
}

// NewBot creates a new Telegram bot
func NewBot(token string) *Bot {
	return &Bot{
		Token:   token,
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

// GetUpdates fetches updates (new messages) from the Telegram API
func (b *Bot) GetUpdates(offset int) ([]Update, error) {
	url := fmt.Sprintf("%sgetUpdates?offset=%d", b.BaseURL, offset)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response struct {
		Ok     bool     `json:"ok"`
		Result []Update `json:"result"`
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	if !response.Ok {
		return nil, fmt.Errorf("failed to fetch updates")
	}

	return response.Result, nil
}

// Update represents an incoming update from the Telegram API
type Update struct {
	UpdateID int    `json:"update_id"`
	Message  Message `json:"message"`
}

// Message represents a Telegram message
type Message struct {
	Chat Chat   `json:"chat"`
	Text string `json:"text"`
}

// Chat represents a Telegram chat
type Chat struct {
	ID int64 `json:"id"`
}
