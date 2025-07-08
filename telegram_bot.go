package main

import (
	"fmt"
	"log"
	"slices"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type telegramBot struct {
	allowedUserIds    []int64
	token             string
	paperlessEndpoint string
	paperlessToken    string
	bot               *tgbotapi.BotAPI
	chats             []*telegramChat
	scanner           *scanner
}

func stringSliceToKeyboard(values []string) tgbotapi.InlineKeyboardMarkup {
	return stringMatrixToKeyboard([][]string{values})
}

func stringMatrixToKeyboard(values [][]string) tgbotapi.InlineKeyboardMarkup {
	var keyboardRows [][]tgbotapi.InlineKeyboardButton
	for _, row := range values {
		keyboardRow := tgbotapi.NewInlineKeyboardRow()
		for _, text := range row {
			button := tgbotapi.NewInlineKeyboardButtonData(text, text)
			keyboardRow = append(keyboardRow, button)
		}
		keyboardRows = append(keyboardRows, keyboardRow)
	}
	return tgbotapi.NewInlineKeyboardMarkup(keyboardRows...)
}

func newTelegramBot(allowedUserIds []int64, token string, scanner *scanner, paperlessEndpoint string, paperlessToken string) (*telegramBot, error) {
	var err error
	bot := telegramBot{
		allowedUserIds:    allowedUserIds,
		token:             token,
		scanner:           scanner,
		paperlessEndpoint: paperlessEndpoint,
		paperlessToken:    paperlessToken,
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
	log.Printf("Allowed users: %v", bot.allowedUserIds)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.bot.GetUpdatesChan(u)

	for update := range updates {
		userId, chatId := getUserAndChatId(update)
		fmt.Printf("Received update from %d\n", userId)
		if slices.Contains(bot.allowedUserIds, userId) {
			fmt.Println("Message from allowed chat")
			chat := bot.getChat(chatId)
			if chat == nil {
				bot.chats = append(bot.chats, newChat(chatId, bot, bot.scanner, bot.paperlessEndpoint, bot.paperlessToken))
				chat = bot.getChat(chatId)
			}
			// Check if we've gotten a message update.
			if update.Message != nil {
				fmt.Printf("Update is message\n")
				// Send the message.
				if chat != nil {
					chat.handleMessage(update.Message)
				}

			} else if update.CallbackQuery != nil {
				fmt.Printf("Update is callback\n")
				if chat != nil {
					chat.handleCallbackQuery(update.CallbackQuery)
				}
				// Respond to the callback query, telling Telegram to show the user
				// a message with the data received.
				callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
				if _, err := bot.bot.Request(callback); err != nil {
					fmt.Println(err.Error())
				}
			}
		}
	}
}

func getUserAndChatId(update tgbotapi.Update) (user int64, chat int64) {
	if update.Message != nil {
		return update.Message.From.ID, update.Message.Chat.ID
	} else if update.CallbackQuery != nil {
		return update.CallbackQuery.From.ID, update.CallbackQuery.Message.Chat.ID
	} else {
		return 0, 0
	}

}

func (bot telegramBot) getChat(chatId int64) *telegramChat {
	for _, chat := range bot.chats {
		if chat.id == chatId {
			return chat
		}
	}
	return nil
}
