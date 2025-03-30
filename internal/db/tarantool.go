package db

import (
	"log"
	"vote-bot/configs"

	"github.com/tarantool/go-tarantool"
)

type ConnTarantool struct {
	conn *tarantool.Connection
}

func Connect() *ConnTarantool {
	cfg := configs.LoadConfig()

	conn, err := tarantool.Connect(
		cfg.Tarantool.Address,
		tarantool.Opts{
			User: cfg.Tarantool.User,
			Pass: cfg.Tarantool.Password,
		})

	if err != nil {
		log.Fatalf("Не удалось подключиться %v", err)
	}
	log.Printf("Успешное подключение к %v", cfg.Tarantool.Address)

	return &ConnTarantool{conn: conn}
}
