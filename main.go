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
	scannerEndpoint := os.Getenv("SCANNER_ENDPOINT")

	fmt.Printf("TELEGRAM_BOT_TOKEN: %s\n", telegramBotToken)
	fmt.Printf("ALLOWED_TELEGRAM_USERS: %s\n", allowedUserIdsString)
	fmt.Printf("SCANNER_ENDPOINT: %s\n", scannerEndpoint)

	var allowedUserIds []int64

	for _, id := range allowedUserIdsString {
		n, err := strconv.ParseInt(id, 10, 64)
		if err == nil {
			allowedUserIds = append(allowedUserIds, n)
		} else {
			fmt.Printf("Failed to parse user %s (error: %s)", id, err)
		}
	}

	scannerFunctions := []scannerFunction{{
		name:   "Einzug Farbe",
		source: adf,
		mode:   color,
	}, {
		name:   "Flachbett Farbe",
		source: flatbed,
		mode:   color,
	}, {
		name:   "Einzug S/W",
		source: adf,
		mode:   gray,
	}, {
		name:   "Flachbett S/W",
		source: flatbed,
		mode:   gray,
	}}

	fmt.Println(allowedUserIds)

	scanner := newScanner(scannerEndpoint, scannerFunctions, "airscan:w1:Samsung C48x Series (SEC30CDA7AA690C)")

	_, err := newTelegramBot(getScannerKeyboard(), allowedUserIds, telegramBotToken, scanner.scan)

	if err != nil {
		log.Panic(err)
	}
}
