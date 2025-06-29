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

	var allowedUserIds []int64

	for _, id := range allowedUserIdsString {
		n, err := strconv.ParseInt(id, 10, 64)
		if err == nil {
			allowedUserIds = append(allowedUserIds, n)
		}
	}

	scannerFunctions := []scannerFunction{{
		name:   "Einzug Farbe",
		source: "ADF",
		mode:   "Color",
	}, {
		name:   "Flachbett Farbe",
		source: "Flatbed",
		mode:   "Color",
	}, {
		name:   "Einzug S/W",
		source: "ADF",
		mode:   "Gray",
	}, {
		name:   "Flachbett S/W",
		source: "Flatbed",
		mode:   "Gray",
	}}

	fmt.Println(allowedUserIds)

	scanner := newScanner(scannerEndpoint, scannerFunctions, "airscan:w1:Samsung C48x Series (SEC30CDA7AA690C)")

	_, err := newTelegramBot(getScannerKeyboard(), allowedUserIds, telegramBotToken, scanner.scan)

	if err != nil {
		log.Panic(err)
	}
}
