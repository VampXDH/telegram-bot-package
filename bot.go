package telegrambot

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

// HandleDocument processes a document sent by the user
func (b *Bot) HandleDocument(update Update) (int, error) {
	if update.Message.Document.FileID == "" {
		return 0, errors.New("no document file ID found")
	}

	fileURL, err := b.GetFileURL(update.Message.Document.FileID)
	if err != nil {
		return 0, err
	}

	lineCount, err := b.CountLinesInFile(fileURL)
	if err != nil {
		return 0, err
	}

	return lineCount, nil
}

// GetFileURL retrieves the file URL from Telegram server
func (b *Bot) GetFileURL(fileID string) (string, error) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/getFile", b.Token)
	data := url.Values{}
	data.Set("file_id", fileID)

	resp, err := b.Client.PostForm(apiURL, data)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("failed to get file URL")
	}

	var fileResponse FileResponse
	err = json.NewDecoder(resp.Body).Decode(&fileResponse)
	if err != nil {
		return "", err
	}

	if !fileResponse.Ok {
		return "", errors.New("failed to get file URL")
	}

	filePath := fileResponse.Result.FilePath
	fileURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", b.Token, filePath)

	return fileURL, nil
}

// CountLinesInFile counts the number of lines in the file from the provided URL
func (b *Bot) CountLinesInFile(fileURL string) (int, error) {
	resp, err := http.Get(fileURL)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, errors.New("failed to download file")
	}

	reader := bufio.NewReader(resp.Body)
	lineCount := 0

	for {
		_, isPrefix, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}

		if !isPrefix {
			lineCount++
		}
	}

	return lineCount, nil
}
