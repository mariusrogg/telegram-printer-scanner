package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {
	telegramBotToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	allowedUserIdsString := strings.Split(os.Getenv("ALLOWED_TELEGRAM_USERS"), ";")

	var allowedUserIds []int64

	for _, id := range allowedUserIdsString {
		n, err := strconv.ParseInt(id, 10, 64)
		if err == nil {
			allowedUserIds = append(allowedUserIds, n)
		}
	}

	fmt.Println(allowedUserIds)

	_, err := newTelegramBot(getScannerKeyboard(), allowedUserIds, telegramBotToken)

	if err != nil {
		log.Panic(err)
	}
}
