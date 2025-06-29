package main

import (
	"io"
	"log"
	"slices"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type messageCallbackType func(message string, callback func(file io.ReadCloser, fileName string))

type telegramBot struct {
	keyboard        tgbotapi.ReplyKeyboardMarkup
	allowedUserIds  []int64
	token           string
	bot             *tgbotapi.BotAPI
	chats           []telegramChat
	messageCallback messageCallbackType
}

func stringSliceToKeyboard(values [][]string) tgbotapi.ReplyKeyboardMarkup {
	var keyboardRows [][]tgbotapi.KeyboardButton
	for _, row := range values {
		keyboardRow := tgbotapi.NewKeyboardButtonRow()
		for _, text := range row {
			button := tgbotapi.NewKeyboardButton(text)
			keyboardRow = append(keyboardRow, button)
		}
		keyboardRows = append(keyboardRows, keyboardRow)
	}
	return tgbotapi.NewReplyKeyboard(keyboardRows...)
}

func getScannerKeyboard() [][]string {
	return [][]string{{
		"Einzug Farbe",
		"Einzug S/W",
		"Einzug Paperless"}, {
		"Flachbett Farbe",
		"Flachbett S/W",
		"Flachbett Paperless"}}
}

func newTelegramBot(keyboard [][]string, allowedUserIds []int64, token string, messageCallback messageCallbackType) (*telegramBot, error) {
	var err error
	bot := telegramBot{
		keyboard:        stringSliceToKeyboard(keyboard),
		allowedUserIds:  allowedUserIds,
		token:           token,
		messageCallback: messageCallback,
	}
	err = bot.initTelegramBot()
	return &bot, err
}

func (bot telegramBot) initTelegramBot() error {
	var err error
	bot.bot, err = tgbotapi.NewBotAPI(bot.token)
	if err == nil {
		bot.run()
	}
	return err
}

func (bot telegramBot) run() {
	bot.bot.Debug = true

	log.Printf("Authorized on account %s", bot.bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.bot.GetUpdatesChan(u)

	for update := range updates {
		// Check if we've gotten a message update.
		if update.Message != nil {
			if slices.Contains(bot.allowedUserIds, update.Message.From.ID) {
				chatId := update.Message.Chat.ID
				chat := bot.getChat(chatId)
				if chat == nil {
					bot.chats = append(bot.chats, *newChat(chatId, bot.keyboard, bot))
					chat = bot.getChat(chatId)
				}
				// Send the message.
				if chat != nil {
					chat.handleMessage(update.Message, bot.messageCallback)
				}
			}
		} else if update.CallbackQuery != nil {
			// Respond to the callback query, telling Telegram to show the user
			// a message with the data received.
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
			if _, err := bot.bot.Request(callback); err != nil {
				panic(err)
			}

			// And finally, send a message containing the data received.
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Data)
			if _, err := bot.bot.Send(msg); err != nil {
				panic(err)
			}
		}
	}
}

func (bot telegramBot) getChat(chatId int64) *telegramChat {
	for _, chat := range bot.chats {
		if chat.id == chatId {
			return &chat
		}
	}
	return nil
}
