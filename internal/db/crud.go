package db

import (
	"log"
	"time"
	"vote-bot/internal/poll"

	"github.com/tarantool/go-tarantool"
)

func (r *ConnTarantool) CreatePoll(p *poll.Poll) error {
	_, err := r.conn.Insert("polls", []interface{}{
		p.ID, p.Creator, p.Question, p.Options, p.Votes, p.Active, p.CreatedAt.Unix(), nil,
	})
	if err != nil {
		log.Printf("Ошибка при сохранении голосования: %v", err)
		return err
	}
	return nil
}

func (r *ConnTarantool) GetPoll(id string) (*poll.Poll, error) {
	resp, err := r.conn.Select("polls", "primary", 0, 1, tarantool.IterEq, []interface{}{id})
	if err != nil {
		log.Printf("Ошибка при получении голосования: %v", err)
		return nil, err
	}

	if len(resp.Tuples()) == 0 {
		return nil, nil
	}

	var p poll.Poll
	tuple := resp.Tuples()[0]
	p.ID = tuple[0].(string)
	p.Creator = tuple[1].(string)
	p.Question = tuple[2].(string)
	optionsRaw := tuple[3].([]interface{})
	options := make([]string, len(optionsRaw))
	for i, opt := range optionsRaw {
		options[i] = opt.(string)
	}
	p.Options = options
	votesRaw := tuple[4].(map[interface{}]interface{})
	votes := make(map[string]int)
	for key, value := range votesRaw {
		keyStr := key.(string)
		valueInt := int(value.(uint64))
		votes[keyStr] = valueInt
	}
	p.Votes = votes
	p.Active = tuple[5].(bool)
	createdAt := time.Unix(int64(tuple[6].(uint64)), 0)
	p.CreatedAt = createdAt

	if tuple[7] != nil {
		closedAt := time.Unix(int64(tuple[7].(uint64)), 0)
		p.ClosedAt = &closedAt
	}

	return &p, nil
}

func (r *ConnTarantool) UpdatePoll(p *poll.Poll) error {
	_, err := r.conn.Replace("polls", []interface{}{
		p.ID, p.Creator, p.Question, p.Options, p.Votes, p.Active, p.CreatedAt.Unix(), nil,
	})
	if err != nil {
		log.Printf("Ошибка при обновлении голосования: %v", err)
		return err
	}
	return nil
}

func (r *ConnTarantool) DeletePoll(id string) error {
	_, err := r.conn.Delete("polls", "primary", []interface{}{id})
	if err != nil {
		log.Printf("Ошибка при удалении голосования: %v", err)
		return err
	}
	return nil
}
