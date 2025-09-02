package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

func main() {
	telegramBotToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	allowedUserIdsString := strings.Split(os.Getenv("ALLOWED_TELEGRAM_USERS"), ";")
	scannerEndpoint := os.Getenv("SCANNER_ENDPOINT")
	scannerDeviceId := os.Getenv("SCANNER_DEVICE_ID")
	paperlessEndpoint := os.Getenv("PAPERLESS_ENDPOINT")
	paperlessToken := os.Getenv("PAPERLESS_TOKEN")

	fmt.Printf("TELEGRAM_BOT_TOKEN: %s\n", telegramBotToken)
	fmt.Printf("ALLOWED_TELEGRAM_USERS: %s\n", allowedUserIdsString)
	fmt.Printf("SCANNER_ENDPOINT: %s\n", scannerEndpoint)
	fmt.Printf("SCANNER_DEVICE_ID: %s\n", scannerDeviceId)
	fmt.Printf("PAPERLESS_ENDPOINT: %s\n", paperlessEndpoint)
	fmt.Printf("PAPERLESS_TOKEN: %s\n", paperlessToken)

	// Disable config dir for pdfcpu
	api.DisableConfigDir()

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
	}, {
		source: adf,
		mode:   gray,
		target: paperless,
	}, {
		source: flatbed,
		mode:   gray,
		target: paperless,
	}, {
		source: adf,
		mode:   color,
		target: paperless,
	}, {
		source: flatbed,
		mode:   color,
		target: paperless,
	}}

	fmt.Println(allowedUserIds)

	scanner := newScanner(scannerEndpoint, scannerFunctions, scannerDeviceId)

	_, err := newTelegramBot(allowedUserIds, telegramBotToken, scanner, paperlessEndpoint, paperlessToken)

	if err != nil {
		log.Panic(err)
	}
}
