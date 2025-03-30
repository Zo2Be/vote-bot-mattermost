package bot

import (
	"fmt"
	"strings"
	"time"
	"vote-bot/internal/db"
	"vote-bot/internal/poll"

	"github.com/mattermost/mattermost-server/v6/model"
)

func NewBot(client *model.Client4, botUser *model.User, channel *model.Channel, dbConn *db.ConnTarantool) *Bot {
	return &Bot{
		client:  client,
		botUser: botUser,
		Channel: channel,
		dbConn:  dbConn,
	}
}

func (b *Bot) HandleMessage(post *model.Post) {
	if post.UserId == b.botUser.Id || post.Message == "" {
		return
	}

	parts := strings.Fields(post.Message)
	if len(parts) == 0 {
		return
	}

	cmd, args := parts[0], parts[1:]
	switch cmd {
	case "!create":
		b.createPoll(post.ChannelId, post.UserId, args)
	case "!vote":
		b.vote(post.ChannelId, args)
	case "!results":
		b.results(post.ChannelId, args)
	case "!close":
		b.closePoll(post.ChannelId, post.UserId, args)
	case "!delete":
		b.deletePoll(post.ChannelId, post.UserId, args)
	case "!help":
		b.SendMessage(post.ChannelId, "Команды: !create, !vote, !results, !close, !delete")
	default:
		b.SendMessage(post.ChannelId, "Неизвестная команда. !help — список доступных команд.")
	}
}

func (b *Bot) createPoll(channelID, userID string, args []string) {
	if len(args) < 2 {
		b.SendMessage(channelID, "Использование: !create <вопрос> | <варианты через запятую>")
		return
	}

	questionWithOptions := strings.Join(args, " ")
	parts := strings.Split(questionWithOptions, "|")
	if len(parts) < 2 {
		b.SendMessage(channelID, "Ошибка: нужен разделитель '|' между вопросом и вариантами ответа!")
		return
	}

	question := strings.TrimSpace(parts[0])
	options := splitOptions(strings.TrimSpace(parts[1]))
	optionSet := make(map[string]bool)
	for _, option := range options {
		if optionSet[option] {
			b.SendMessage(channelID, "Ошибка: Варианты ответа не должны повторяться!")
			return
		}
		optionSet[option] = true
	}
	pollID := fmt.Sprintf("poll-%d", time.Now().Unix())

	newPoll := &poll.Poll{
		ID:        pollID,
		Creator:   userID,
		Question:  question,
		Options:   options,
		Votes:     make(map[string]int),
		Active:    true,
		CreatedAt: time.Now(),
	}

	if err := b.dbConn.CreatePoll(newPoll); err != nil {
		b.SendMessage(channelID, "Ошибка при сохранении голосования в БД")
		return
	}

	optionsText := strings.Join(options, ", ")
	b.SendMessage(channelID, fmt.Sprintf("Голосование создано: *%s*\nID: %s\nВарианты: %s", question, pollID, optionsText))
}

func (b *Bot) vote(channelID string, args []string) {
	if len(args) < 2 {
		b.SendMessage(channelID, "Использование: !vote <poll_id> <вариант>")
		return
	}

	getPoll, err := b.dbConn.GetPoll(args[0])
	if err != nil || getPoll == nil || !getPoll.Active {
		b.SendMessage(channelID, "Голосование не найдено или закрыто")
		return
	}

	validOption := false
	for _, option := range getPoll.Options {
		if option == args[1] {
			validOption = true
			break
		}
	}
	if !validOption {
		b.SendMessage(channelID, fmt.Sprintf("Вариант '%s' не найден в голосовании!", args[1]))
		return
	}

	getPoll.Votes[args[1]]++
	b.dbConn.UpdatePoll(getPoll)
	b.SendMessage(channelID, fmt.Sprintf("Голос за '%s' учтен!", args[1]))
}

func (b *Bot) results(channelID string, args []string) {
	if len(args) < 1 {
		b.SendMessage(channelID, "Использование: !results <poll_id>")
		return
	}

	getPoll, err := b.dbConn.GetPoll(args[0])
	if err != nil {
		b.SendMessage(channelID, "Ошибка при получении голосования")
		return
	}

	if getPoll == nil {
		b.SendMessage(channelID, "Голосование не найдено")
		return
	}

	results := fmt.Sprintf("Результаты голосования '%s':\n", getPoll.Question)
	for _, opt := range getPoll.Options {
		count := getPoll.Votes[opt]
		results += fmt.Sprintf("%s: %d голосов\n", opt, count)
	}
	b.SendMessage(channelID, results)
}

func (b *Bot) closePoll(channelID, userID string, args []string) {
	if len(args) < 1 {
		b.SendMessage(channelID, "Использование: !close <poll_id>")
		return
	}

	getPoll, err := b.dbConn.GetPoll(args[0])
	if err != nil {
		b.SendMessage(channelID, "Ошибка при получении голосования")
		return
	}

	if getPoll == nil {
		b.SendMessage(channelID, "Голосование не найдено или уже удалено")
		return
	}

	if getPoll.Creator != userID {
		b.SendMessage(channelID, "Вы не можете закрыть это голосование")
		return
	}

	getPoll.Active = false
	b.dbConn.UpdatePoll(getPoll)
	b.SendMessage(channelID, fmt.Sprintf("Голосование %s закрыто", args[0]))
}

func (b *Bot) deletePoll(channelID, userID string, args []string) {
	if len(args) < 1 {
		b.SendMessage(channelID, "Использование: !delete <poll_id>")
		return
	}

	getPoll, err := b.dbConn.GetPoll(args[0])
	if err != nil {
		b.SendMessage(channelID, "Ошибка при получении голосования")
		return
	}

	if getPoll == nil {
		b.SendMessage(channelID, "Голосование не найдено или уже удалено")
		return
	}

	if getPoll.Creator != userID {
		b.SendMessage(channelID, "Вы не можете удалить это голосование")
		return
	}

	b.dbConn.DeletePoll(args[0])
	b.SendMessage(channelID, fmt.Sprintf("Голосование %s удалено", args[0]))
}

func (b *Bot) SendMessage(channelID, msg string) {
	b.client.CreatePost(&model.Post{ChannelId: channelID, Message: msg})
}

func splitOptions(options string) []string {
	parts := strings.Split(options, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}
