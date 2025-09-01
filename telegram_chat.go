package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

type ChatState int

const (
	stateInit            ChatState = iota
	stateUseLast         ChatState = iota
	stateTarget          ChatState = iota
	stateSource          ChatState = iota
	stateDuplex          ChatState = iota
	stateMode            ChatState = iota
	stateScanDuplexFront ChatState = iota
	stateScanDuplexRear  ChatState = iota
	stateScanSimple      ChatState = iota
)

var chatState = map[ChatState]string{
	stateInit:            "stateInit",
	stateUseLast:         "stateUseLast",
	stateTarget:          "stateTarget",
	stateSource:          "stateSource",
	stateDuplex:          "stateDuplex",
	stateMode:            "stateMode",
	stateScanDuplexFront: "stateScanDuplexFront",
	stateScanDuplexRear:  "stateScanDuplexRear",
	stateScanSimple:      "stateScanSimple",
}

func (cs ChatState) String() string {
	return chatState[cs]
}

type Decision string

const (
	yes Decision = "Yes"
	no  Decision = "No"
)

var decision = map[Decision]string{
	yes: string(yes),
	no:  string(no),
}

func (d Decision) String() string {
	return decision[d]
}

type telegramChat struct {
	id                int64
	bot               telegramBot
	scanner           *scanner
	state             ChatState
	paperlessEndpoint string
	paperlessToken    string
	currentTarget     ScannerTarget
	currentSource     ScannerSource
	currentMode       ScannerMode
	currentDuplex     Decision
	currentMessage    tgbotapi.Message
	currentFunction   ScannerFunction
	duplexFrontFile   io.ReadSeeker
}

func newChat(id int64, bot telegramBot, scanner *scanner, paperlessEndpoint string, paperlessToken string) *telegramChat {
	return &telegramChat{
		id:                id,
		bot:               bot,
		scanner:           scanner,
		state:             stateInit,
		paperlessEndpoint: paperlessEndpoint,
		paperlessToken:    paperlessToken,
	}
}

func (chat telegramChat) sendFile(file io.ReadCloser, fileName string) error {
	tgFile := tgbotapi.FileReader{
		Name:   fileName,
		Reader: file,
	}
	defer file.Close()

	_, err := chat.bot.bot.SendMediaGroup(tgbotapi.NewMediaGroup(chat.id, []interface{}{
		tgbotapi.NewInputMediaDocument(tgFile),
	}))
	return err
}

// Converts a slice of any type that implements fmt.Stringer to a slice of strings.
func sliceToStringSlice[T fmt.Stringer](input []T) []string {
	output := make([]string, 0, len(input))
	for _, v := range input {
		output = append(output, v.String())
	}
	return output
}

func (chat *telegramChat) handleMessage(message *tgbotapi.Message) {
	fmt.Println("Message received: " + message.Text)
	fmt.Println("Current state: " + chat.state.String())
	if message.Text == "/restart" || chat.state == stateInit {
		chat.runInit()
	}
}

func (chat *telegramChat) handleCallbackQuery(callbackQuery *tgbotapi.CallbackQuery) {
	fmt.Println("Message received: " + callbackQuery.Data)
	fmt.Println("Current state: " + chat.state.String())
	switch chat.state {
	case stateInit:
		chat.runInit()

	case stateUseLast:
		if Decision(callbackQuery.Data) == yes {
			chat.chooseScanState()
		} else {
			chat.prepStateTarget()
		}

	case stateTarget:
		chat.currentTarget = ScannerTarget(callbackQuery.Data)
		chat.prepStateSource()

	case stateSource:
		chat.currentSource = ScannerSource(callbackQuery.Data)
		if chat.currentSource == adf {
			chat.prepStateDuplex()
		} else {
			chat.prepStateMode()
		}

	case stateDuplex:
		chat.currentDuplex = Decision(callbackQuery.Data)
		chat.prepStateMode()

	case stateMode:
		chat.currentMode = ScannerMode(callbackQuery.Data)
		chat.chooseScanState()

	case stateScanDuplexFront:
		if Decision(callbackQuery.Data) == yes {
			file, _, err := chat.currentFunction.scan(chat.scanner.endpoint, chat.scanner.deviceId)
			if err != nil {
				fmt.Printf("failed to scan: %s\n", err.Error())
				chat.prepStateUseLast()
			} else if chat.currentDuplex == yes {
				chat.duplexFrontFile = readerToReadSeeker(file)
				chat.prepStateScanDuplexRear()
			} else {
				chat.prepStateUseLast()
			}
		} else {
			chat.prepStateUseLast()
		}
	case stateScanDuplexRear:
		if Decision(callbackQuery.Data) == yes {
			file, filename, err := chat.currentFunction.scan(chat.scanner.endpoint, chat.scanner.deviceId)
			if err != nil {
				fmt.Printf("failed to scan: %s\n", err.Error())
				chat.prepStateUseLast()
				break
			}
			front := chat.duplexFrontFile
			rear := readerToReadSeeker(file)
			frontPages, err := getPages(front)
			if err != nil {
				chat.prepStateUseLast()
				break
			}
			rearPages, err := getPages(rear)
			if err != nil {
				chat.prepStateUseLast()
				break
			}
			err = chat.finish(chat.currentTarget, orderAndMerge(frontPages, rearPages), filename)
			chat.prepStateUseLast()
			if err != nil {
				fmt.Println(err)
			}
		} else {
			chat.prepStateUseLast()
		}
	case stateScanSimple:
		if Decision(callbackQuery.Data) == yes {
			file, filename, err := chat.currentFunction.scan(chat.scanner.endpoint, chat.scanner.deviceId)
			if err != nil {
				fmt.Printf("failed to scan: %s\n", err.Error())
			} else {
				err = chat.finish(chat.currentTarget, file, filename)
				chat.prepStateUseLast()
				if err != nil {
					fmt.Println(err)
				}
			}
		}
		chat.prepStateUseLast()

	default:
		fmt.Printf("Chat state %s is unknown", chat.state)
	}

}

func (chat *telegramChat) deleteLastMessage() {
	if chat.currentMessage.MessageID != 0 {
		deleteMessage := tgbotapi.NewDeleteMessage(chat.id, chat.currentMessage.MessageID)
		chat.bot.bot.Send(deleteMessage)
		chat.currentMessage.MessageID = 0
	}
}

func (chat *telegramChat) runInit() {
	chat.deleteLastMessage()
	if chat.currentTarget != "" && chat.currentSource != "" && chat.currentMode != "" {
		chat.prepStateUseLast()
	} else {
		chat.prepStateTarget()
	}
}

func (chat *telegramChat) finish(target ScannerTarget, file io.ReadCloser, fileName string) error {
	switch target {
	case telegram:
		return chat.sendFile(file, fileName)
	case paperless:

		url := chat.paperlessEndpoint + "/api/documents/post_document/"
		method := "POST"

		payload := &bytes.Buffer{}
		writer := multipart.NewWriter(payload)
		err := writer.WriteField("from_webui", "false")
		if err != nil {
			return err
		}
		defer file.Close()
		part, err := writer.CreateFormFile("document", fileName)
		if err != nil {
			return err
		}
		_, err = io.Copy(part, file)
		if err != nil {
			return err
		}
		err = writer.Close()
		if err != nil {
			fmt.Println(err)
			return err
		}

		client := &http.Client{}
		req, err := http.NewRequest(method, url, payload)

		if err != nil {
			fmt.Println(err)
			return err
		}

		req.Header.Add("accept", "application/json")
		req.Header.Add("Authorization", "Token "+chat.paperlessToken)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		res, err := client.Do(req)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}
		fmt.Println(string(body))
		return nil
	}
	return fmt.Errorf("target not supported")
}

func orderAndMerge(front []*api.PageSpan, rear []*api.PageSpan) io.ReadCloser {
	if front == nil || rear == nil {
		fmt.Println("Front or rear pages are nil")
		return nil
	}
	if len(front) != len(rear) {
		fmt.Printf("Different number of front (%d) and rear pages(%d)\n", len(front), len(rear))
		return nil
	}
	pages := []io.ReadSeeker{}
	for i := 0; i < len(front); i++ {
		frontPage := front[i]
		rearPage := rear[len(rear)-i-1]
		if frontPage == nil || rearPage == nil {
			fmt.Printf("Nil page found at index %d\n", i)
			return nil
		}
		pages = append(pages, readerToReadSeeker(frontPage.Reader), readerToReadSeeker(rearPage.Reader))
	}
	reader, writer := io.Pipe()
	go func() {
		defer writer.Close()
		err := api.MergeRaw(pages, writer, false, model.NewDefaultConfiguration())
		if err != nil {
			fmt.Printf("failed to merge pdfs: %s\n", err.Error())
			writer.CloseWithError(err)
		}
	}()
	return reader
}

func readerToReadSeeker(file io.Reader) io.ReadSeeker {
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("failed to read bytes from file: %s\n", err.Error())
	}
	return bytes.NewReader(fileBytes)
}

func getPages(file io.ReadSeeker) ([]*api.PageSpan, error) {
	pages, err := api.SplitRaw(file, 1, model.NewDefaultConfiguration())
	if err != nil {
		fmt.Printf("failed to split pdf: %s\n", err.Error())
	}
	return pages, err
}

func (chat *telegramChat) updateMessage(text string, replyMarkup tgbotapi.InlineKeyboardMarkup) (tgbotapi.Message, error) {
	messageConfig := tgbotapi.NewEditMessageTextAndMarkup(chat.id, chat.currentMessage.MessageID, text, replyMarkup)
	return chat.bot.bot.Send(messageConfig)
}

func (chat *telegramChat) chooseScanState() {
	if chat.currentSource == adf && chat.currentDuplex == yes {
		chat.prepStateScanDuplexFront()
	} else {
		chat.prepStateScanSimple()
	}
}

func (chat *telegramChat) prepStateUseLast() {
	chat.deleteLastMessage()
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Use last configuration:\nTarget: %s\nSource: %s\nMode: %s\n", chat.currentTarget, chat.currentSource, chat.currentMode))
	if chat.currentMode == ScannerMode(adf) {
		builder.WriteString(fmt.Sprintf("Duplex: %s", chat.currentDuplex))
	}
	prepState(chat, stateUseLast, []fmt.Stringer{yes, no}, builder.String(), true)
}
func (chat *telegramChat) prepStateTarget() {
	prepState(chat, stateTarget, chat.scanner.getTargets(), "Select a target to scan to", chat.currentMessage.MessageID == 0)
}

func (chat *telegramChat) prepStateSource() {
	prepState(chat, stateSource, chat.scanner.getSources(chat.currentTarget), "Select a source", false)
}
func (chat *telegramChat) prepStateMode() {
	prepState(chat, stateMode, chat.scanner.getModes(chat.currentTarget, ScannerSource(chat.currentSource)), "Select a scan mode", false)
}

func (chat *telegramChat) prepStateDuplex() {
	prepState(chat, stateDuplex, []fmt.Stringer{yes, no}, "Duplex scan?", false)
}

func (chat *telegramChat) prepStateScanDuplexFront() {
	prepState(chat, stateScanDuplexFront, []fmt.Stringer{yes, no}, "Start front scan?", false)
	chat.currentFunction = *chat.scanner.getFunction(chat.currentTarget, chat.currentSource, chat.currentMode)
}
func (chat *telegramChat) prepStateScanDuplexRear() {
	prepState(chat, stateScanDuplexRear, []fmt.Stringer{yes, no}, "Start rear scan?", false)
	chat.currentFunction = *chat.scanner.getFunction(chat.currentTarget, chat.currentSource, chat.currentMode)
}

func (chat *telegramChat) prepStateScanSimple() {
	prepState(chat, stateScanSimple, []fmt.Stringer{yes, no}, "Start scan?", false)
	chat.currentFunction = *chat.scanner.getFunction(chat.currentTarget, chat.currentSource, chat.currentMode)
}

func prepState[T fmt.Stringer](chat *telegramChat, state ChatState, slice []T, message string, init bool) {
	keyboard := stringSliceToKeyboard(sliceToStringSlice(slice))
	if init {
		var err error
		answer := tgbotapi.NewMessage(chat.id, message)
		answer.ReplyMarkup = keyboard
		chat.currentMessage, err = chat.bot.bot.Send(answer)
		if err != nil {
			fmt.Printf("Failed to send message: %s\n", err.Error())
			return
		}
	} else {
		_, err := chat.updateMessage(message, keyboard)
		if err != nil {
			fmt.Printf("Failed to send message: %s\n", err.Error())
			return
		}
	}
	chat.state = state
}
