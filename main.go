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
	scannerDeviceId := os.Getenv("SCANNER_DEVICE_ID")

	fmt.Printf("TELEGRAM_BOT_TOKEN: %s\n", telegramBotToken)
	fmt.Printf("ALLOWED_TELEGRAM_USERS: %s\n", allowedUserIdsString)
	fmt.Printf("SCANNER_ENDPOINT: %s\n", scannerEndpoint)
	fmt.Printf("SCANNER_DEVICE_ID: %s\n", scannerDeviceId)

	var allowedUserIds []int64

	for _, id := range allowedUserIdsString {
		n, err := strconv.ParseInt(id, 10, 64)
		if err == nil {
			allowedUserIds = append(allowedUserIds, n)
		} else {
			fmt.Printf("Failed to parse user %s (error: %s)", id, err)
		}
	}

	scannerFunctions := []ScannerFunction{{
		source: adf,
		mode:   color,
		target: telegram,
	}, {
		source: flatbed,
		mode:   color,
		target: telegram,
	}, {
		source: adf,
		mode:   gray,
		target: telegram,
	}, {
		source: flatbed,
		mode:   gray,
		target: telegram,
	}}

	fmt.Println(allowedUserIds)

	scanner := newScanner(scannerEndpoint, scannerFunctions, scannerDeviceId)

	_, err := newTelegramBot(getScannerKeyboard(), allowedUserIds, telegramBotToken, scanner)

	if err != nil {
		log.Panic(err)
	}
}
