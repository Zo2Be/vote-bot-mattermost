package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"
	"vote-bot/internal/db"

	"vote-bot/configs"
	"vote-bot/internal/bot"

	"github.com/mattermost/mattermost-server/v6/model"
)

func main() {
	config := configs.LoadConfig()

	client := model.NewAPIv4Client(config.Mattermost.ServerURL.String())
	client.SetToken(config.Mattermost.Token)

	botUser, _, err := client.GetUserByUsername(config.Mattermost.UserName, "")
	if err != nil {
		log.Fatalf("Ошибка при получении пользователя: %v", err)
	}

	team, _, err := client.GetTeamByName(config.Mattermost.TeamName, "")
	if err != nil {
		log.Fatalf("Ошибка при получении команды: %v", err)
	}

	channel, _, err := client.GetChannelByName(config.Mattermost.Channel, team.Id, "")
	if err != nil {
		log.Fatalf("Ошибка при получении канала: %v", err)
	}
	dbConn := db.Connect()
	b := bot.NewBot(client, botUser, channel, dbConn)

	b.SendMessage(channel.Id, "Привет! Я бот для голосования.\n"+
		"Вот мои команды:\n"+
		"• !create <вопрос> | <варианты через запятую> — создать голосование\n"+
		"• !vote <poll_id> <вариант> — проголосовать\n"+
		"• !results <poll_id> — посмотреть результаты\n"+
		"• !close <poll_id> — завершить голосование (только создатель)\n"+
		"• !delete <poll_id> — удалить голосование (только создатель)\n"+
		"• !help — список команд")
	log.Println("Бот успешно запущен!")

	wsURL := fmt.Sprintf("ws://%s/api/v4/websocket?token=%s", config.Mattermost.ServerURL.Host, config.Mattermost.Token)

	setupGracefulShutdown()
	listenToWebSocket(wsURL, client, channel, b, botUser)
}

func setupGracefulShutdown() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		log.Println("Выключение бота...")
		os.Exit(0)
	}()
}

func listenToWebSocket(wsURL string, client *model.Client4, channel *model.Channel, b *bot.Bot, botUser *model.User) {
	for {
		wsClient, err := model.NewWebSocketClient4(wsURL, client.AuthToken)
		if err != nil {
			log.Printf("Ошибка WebSocket: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		log.Println("WebSocket подключен")
		wsClient.Listen()

		for event := range wsClient.EventChannel {
			go handleWebSocketEvent(event, b, botUser)
		}

		log.Println("WebSocket разорван, переподключение через 5 секунд...")
		time.Sleep(5 * time.Second)
	}
}

func handleWebSocketEvent(event *model.WebSocketEvent, b *bot.Bot, botUser *model.User) {
	postData, ok := event.GetData()["post"].(string)
	if !ok {
		return
	}

	post := &model.Post{}
	if err := json.Unmarshal([]byte(postData), post); err != nil {
		log.Printf("Ошибка разбора JSON: %v", err)
		return
	}

	if post.UserId == botUser.Id {
		return
	}

	b.HandleMessage(post)
}
