package telegrambot

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Bot struct {
	Token    string
	Client   *http.Client
	commands map[string]func(chatID int64) string
	mu       sync.Mutex
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

// SendFile sends a file to a specific chat ID
func (b *Bot) SendFile(chatID int64, filePath string, caption string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendDocument", b.Token)

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("could not open file: %v", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("document", filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("could not create form file: %v", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return fmt.Errorf("could not copy file content: %v", err)
	}

	writer.WriteField("chat_id", fmt.Sprintf("%d", chatID))
	if caption != "" {
		writer.WriteField("caption", caption)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("could not close writer: %v", err)
	}

	req, err := http.NewRequest("POST", apiURL, body)
	if err != nil {
		return fmt.Errorf("could not create request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := b.Client.Do(req)
	if err != nil {
		return fmt.Errorf("could not send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send file: %v", resp.Status)
	}

	return nil
}

// Update represents an update from Telegram
type Update struct {
	UpdateID int     `json:"update_id"`
	Message  Message `json:"message"`
}

// Message represents a message from Telegram
type Message struct {
	MessageID int      `json:"message_id"`
	From      User     `json:"from"`
	Chat      Chat     `json:"chat"`
	Date      int      `json:"date"`
	Text      string   `json:"text"`
	Document  Document `json:"document"`  // Field untuk dokumen yang dikirim
}

// Document represents a document sent to the bot
type Document struct {
	FileID   string `json:"file_id"`
	FileName string `json:"file_name"`
	FileSize int    `json:"file_size"`
}

// User represents a user on Telegram
type User struct {
	ID           int    `json:"id"`
	IsBot        bool   `json:"is_bot"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Username     string `json:"username"`
	LanguageCode string `json:"language_code"`
}

// Chat represents a chat on Telegram
type Chat struct {
	ID        int64  `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// UpdateResponse represents the response from Telegram getUpdates method
type UpdateResponse struct {
	Ok     bool     `json:"ok"`
	Result []Update `json:"result"`
}

// FileResponse represents the response from Telegram getFile method
type FileResponse struct {
	Ok     bool   `json:"ok"`
	Result File   `json:"result"`
}

// File represents the file information from Telegram getFile response
type File struct {
	FileID   string `json:"file_id"`
	FilePath string `json:"file_path"`
}
