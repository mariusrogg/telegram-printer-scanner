package main

import (
	"fmt"
	"io"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type telegramChat struct {
	id          int64
	replyMarkup interface{}
	bot         telegramBot
}

func newChat(id int64, replyMarkup interface{}, bot telegramBot) *telegramChat {
	return &telegramChat{
		id:          id,
		replyMarkup: replyMarkup,
		bot:         bot,
	}
}

func (chat telegramChat) sendFile(file io.ReadCloser, fileName string) {
	tgFile := tgbotapi.FileReader{
		Name:   fileName,
		Reader: file,
	}

	chat.bot.bot.SendMediaGroup(tgbotapi.NewMediaGroup(chat.id, []interface{}{
		tgbotapi.NewInputMediaDocument(tgFile),
	}))
}

func (chat telegramChat) handleMessage(message *tgbotapi.Message, messageCallback messageCallbackType) {
	fmt.Println("Message received: " + message.Text)
	messageCallback(message.Text, chat.sendFile)
}
