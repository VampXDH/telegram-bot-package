package telegrambot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

// Bot struct untuk menyimpan token bot
type Bot struct {
	Token string
}

// Update struct untuk mem-parsing pembaruan yang diterima
type Update struct {
	UpdateID int           `json:"update_id"`
	Message  MessageStruct `json:"message"`
}

// MessageStruct untuk mem-parsing pesan
type MessageStruct struct {
	MessageID int      `json:"message_id"`
	From      User     `json:"from"`
	Chat      Chat     `json:"chat"`
	Date      int      `json:"date"`
	Text      string   `json:"text"`
	Entities  []Entity `json:"entities"`
}

// User struct untuk mem-parsing informasi pengguna
type User struct {
	ID        int    `json:"id"`
	IsBot     bool   `json:"is_bot"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
}

// Chat struct untuk mem-parsing informasi chat
type Chat struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}

// Entity struct untuk mem-parsing entitas pesan
type Entity struct {
	Offset int    `json:"offset"`
	Length int    `json:"length"`
	Type   string `json:"type"`
}

// NewBot membuat instance baru dari Bot
func NewBot(token string) *Bot {
	return &Bot{
		Token: token,
	}
}

// SendMessage mengirim pesan ke chat tertentu
func (b *Bot) SendMessage(chatID int64, text string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", b.Token)
	data := url.Values{}
	data.Set("chat_id", strconv.FormatInt(chatID, 10))
	data.Set("text", text)

	resp, err := http.PostForm(apiURL, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("failed to send message: %s", string(bodyBytes))
	}

	return nil
}

// GetUpdates mengambil pembaruan baru dari API Telegram
func (b *Bot) GetUpdates(offset int) ([]Update, error) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates", b.Token)
	data := url.Values{}
	data.Set("offset", strconv.Itoa(offset))

	resp, err := http.PostForm(apiURL, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get updates: %s", string(bodyBytes))
	}

	var updatesResp struct {
		Ok     bool     `json:"ok"`
		Result []Update `json:"result"`
	}
	err = json.NewDecoder(resp.Body).Decode(&updatesResp)
	if err != nil {
		return nil, err
	}

	return updatesResp.Result, nil
}
