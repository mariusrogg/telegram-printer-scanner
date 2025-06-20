package main

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type telegramChat struct {
	id          int64
	replyMarkup interface{}
}

func newChat(id int64, replyMarkup interface{}) *telegramChat {
	return &telegramChat{
		id:          id,
		replyMarkup: replyMarkup,
	}
}

func (chat telegramChat) createMessage(text string) *tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(chat.id, text)
	msg.ReplyMarkup = chat.replyMarkup

	return &msg
}
