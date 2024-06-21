package main

// сюда писать код

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	tgbotapi "github.com/skinass/telegram-bot-api/v5"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

var (
	// @BotFather в телеграме даст вам это
	BotToken = getBotToken()

	// урл выдаст вам игрок или хероку
	WebhookURL = "https://09692455be5d3e9a5fd66c2e610b512b.serveo.net"

	mu = &sync.RWMutex{}

	commandHandlers = map[string]func(tgbotapi.Update) []tgbotapi.MessageConfig{
		"/tasks":     taskHandler,
		"/new":       newTaskHandler,
		"/assign_":   assignHandler,
		"/unassign_": unassignHandler,
		"/resolve_":  resolveHandler,
		"/my":        myHandler,
		"/owner":     ownerHandler,
	}
)

func getBotToken() string {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	token := os.Getenv("BotToken")
	if token == "" {
		fmt.Println("ERROR: BotToken environment variable is not set.")
		os.Exit(1)
	}
	return token
}

const (
	taska = `Задача "`
	doggy = " by @"
)

var nextID = 1

func generateID() int {
	id := nextID
	nextID++
	return id
}

type task struct {
	id             int
	text           string
	userMade       string
	userMadeChatID int64
	userExec       string
	userExecChatID int64
}

var tasksList []task

func taskHandler(update tgbotapi.Update) []tgbotapi.MessageConfig {
	mu.RLock()
	defer mu.RUnlock()
	if len(tasksList) == 0 {
		msg := tgbotapi.NewMessage(
			update.Message.Chat.ID,
			"Нет задач",
		)
		return []tgbotapi.MessageConfig{msg}
	}

	var text string

	for _, v := range tasksList {
		text += strconv.Itoa(v.id) + ". " + v.text + doggy + v.userMade + "\n"
		switch v.userExec {
		case "":
			text += "/assign_" + strconv.Itoa(v.id)
		case update.Message.From.UserName:
			text += "assignee: я" + "\n" + "/unassign_" + strconv.Itoa(v.id) + " /resolve_" + strconv.Itoa(v.id)
		default:
			text += "assignee: @" + v.userExec
		}
		text += "\n\n"
	}
	return []tgbotapi.MessageConfig{tgbotapi.NewMessage(
		update.Message.Chat.ID,
		text[:len(text)-2],
	)}
}

func newTaskHandler(update tgbotapi.Update) []tgbotapi.MessageConfig {
	newTask := task{
		id:             generateID(),
		text:           update.Message.Text[5:],
		userMade:       update.Message.From.UserName,
		userMadeChatID: update.Message.Chat.ID,
	}
	mu.Lock()
	tasksList = append(tasksList, newTask)
	mu.Unlock()
	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		taska+newTask.text+`" создана, id=`+strconv.Itoa(newTask.id),
	)
	return []tgbotapi.MessageConfig{msg}
}

func notificationHandler(update tgbotapi.Update, v task, typeOf string) []tgbotapi.MessageConfig {
	var messages []tgbotapi.MessageConfig

	switch typeOf {
	case "assign":
		if v.userExec == "" {
			msg := tgbotapi.NewMessage(
				v.userMadeChatID,
				taska+v.text+`" назначена на @`+update.Message.From.UserName,
			)
			messages = append(messages, msg)
		} else {
			msg := tgbotapi.NewMessage(
				v.userExecChatID,
				taska+v.text+`" назначена на @`+update.Message.From.UserName,
			)
			messages = append(messages, msg)
		}

		msg := tgbotapi.NewMessage(
			update.Message.Chat.ID,
			taska+v.text+`" назначена на вас`,
		)
		messages = append(messages, msg)
	case "unassign":
		msg := tgbotapi.NewMessage(
			update.Message.Chat.ID,
			"Принято",
		)
		messages = append(messages, msg)
		msg = tgbotapi.NewMessage(
			v.userMadeChatID,
			taska+v.text+`" осталась без исполнителя`,
		)
		messages = append(messages, msg)
	}

	return messages
}

func assignHandler(update tgbotapi.Update) []tgbotapi.MessageConfig {
	taskID, err := strconv.Atoi(update.Message.Text[8:])
	if err != nil {
		msg := tgbotapi.NewMessage(
			update.Message.Chat.ID,
			err.Error(),
		)
		return []tgbotapi.MessageConfig{msg}
	}

	mu.Lock()
	defer mu.Unlock()
	for k, v := range tasksList {
		if v.id == taskID {
			tasksList[k].userExec = update.Message.From.UserName
			tasksList[k].userExecChatID = update.Message.Chat.ID

			return notificationHandler(update, v, "assign")
		}
	}

	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		"Задача с таким ID не найдена",
	)
	return []tgbotapi.MessageConfig{msg}
}

func unassignHandler(update tgbotapi.Update) []tgbotapi.MessageConfig {
	taskID, err := strconv.Atoi(update.Message.Text[10:])
	if err != nil {
		msg := tgbotapi.NewMessage(
			update.Message.Chat.ID,
			err.Error(),
		)
		return []tgbotapi.MessageConfig{msg}
	}
	mu.Lock()
	defer mu.Unlock()
	for k, v := range tasksList {
		if v.id == taskID {
			if v.userExec != update.Message.From.UserName {
				msg := tgbotapi.NewMessage(
					update.Message.Chat.ID,
					"Задача не на вас",
				)
				return []tgbotapi.MessageConfig{msg}
			}

			tasksList[k].userExec = ""

			return notificationHandler(update, v, "unassign")
		}
	}

	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		"Задача с таким ID не найдена",
	)
	return []tgbotapi.MessageConfig{msg}
}

func resolveHandler(update tgbotapi.Update) []tgbotapi.MessageConfig {
	var messages []tgbotapi.MessageConfig
	taskID, err := strconv.Atoi(update.Message.Text[9:])
	if err != nil {
		msg := tgbotapi.NewMessage(
			update.Message.Chat.ID,
			err.Error(),
		)
		messages = append(messages, msg)
		return messages
	}
	mu.Lock()
	defer mu.Unlock()
	for k, v := range tasksList {
		if v.id == taskID {
			if v.userExec != update.Message.From.UserName {
				msg := tgbotapi.NewMessage(
					v.userMadeChatID,
					"Задача не на вас",
				)
				messages = append(messages, msg)
				return messages
			}

			msg := tgbotapi.NewMessage(
				v.userMadeChatID,
				taska+v.text+`" выполнена @`+v.userExec,
			)
			messages = append(messages, msg)
			msg = tgbotapi.NewMessage(
				update.Message.Chat.ID,
				taska+v.text+`" выполнена`,
			)
			messages = append(messages, msg)

			tasksList = append(tasksList[:k], tasksList[k+1:]...)

			return messages
		}
	}

	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		"Задача с таким ID не найдена",
	)
	messages = append(messages, msg)
	return messages
}

func sendResponse(update tgbotapi.Update, text string) tgbotapi.MessageConfig {
	if text != "" {
		msg := tgbotapi.NewMessage(
			update.Message.Chat.ID,
			text[:len(text)-2],
		)
		return msg
	}

	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		"У вас нет назначенных задач",
	)
	return msg
}

func myHandler(update tgbotapi.Update) []tgbotapi.MessageConfig {
	var messages []tgbotapi.MessageConfig
	var text string
	mu.RLock()
	for _, v := range tasksList {
		if v.userMade == update.Message.From.UserName {
			text += strconv.Itoa(v.id) + ". " + v.text + doggy + v.userMade + "\n" +
				"/unassign_" + strconv.Itoa(v.id) + " /resolve_" + strconv.Itoa(v.id) + "\n\n"
		}
	}
	mu.RUnlock()

	messages = append(messages, sendResponse(update, text))
	return messages
}

func ownerHandler(update tgbotapi.Update) []tgbotapi.MessageConfig {
	var messages []tgbotapi.MessageConfig
	var text string
	mu.RLock()
	for _, v := range tasksList {
		if v.userMade == update.Message.From.UserName {
			text += strconv.Itoa(v.id) + ". " + v.text + doggy + v.userMade + "\n" +
				"/assign_" + strconv.Itoa(v.id) + "\n\n"
		}
	}
	mu.RUnlock()

	messages = append(messages, sendResponse(update, text))
	return messages
}

func updateHandler(update tgbotapi.Update) []tgbotapi.MessageConfig {
	if update.Message == nil {
		return nil // Нет сообщения для обработки
	}

	text := update.Message.Text
	for cmd, handler := range commandHandlers {
		if strings.HasPrefix(text, cmd) {
			return handler(update)
		}
	}

	return nil
}

func startTaskBot(ctx context.Context) error {

	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		log.Printf("NewBotAPI failed: %s", err)
		return err
	}

	bot.Debug = true
	fmt.Printf("Authorized on account %s\n", bot.Self.UserName)

	wh, err := tgbotapi.NewWebhook(WebhookURL)
	if err != nil {
		log.Printf("NewWebhook failed: %s", err)
		return err
	}

	_, err = bot.Request(wh)
	if err != nil {
		log.Printf("SetWebhook failed: %s", err)
		return err
	}

	updates := bot.ListenForWebhook("/")

	http.HandleFunc("/state", func(w http.ResponseWriter, r *http.Request) {
		_, err = w.Write([]byte("all is working"))
		if err != nil {
			return
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	go func() {
		log.Fatalln("http err:", http.ListenAndServe(":"+port, nil))
	}()
	fmt.Println("start listen :" + port)

	for {
		select {
		case update := <-updates:
			log.Printf("upd: %#v\n", update)
			messages := updateHandler(update)
			for _, v := range messages {
				_, err = bot.Send(v)
				if err != nil {
					return err
				}
			}
		case <-ctx.Done():

			// Возвращаем nil, чтобы не считать это ошибкой
			// Сделано, чтобы тест не ругался на graceful shutdown
			if ctx.Err() == context.Canceled {
				log.Println("Operation was canceled")
				return nil
			}
			return ctx.Err()
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cancel()
	}()

	err := startTaskBot(ctx)
	if err != nil {
		log.Println(err)
	}

	log.Println("Bot stopped gracefully")
}
