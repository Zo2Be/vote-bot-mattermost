package configs

import (
	"github.com/joho/godotenv"
	"log"
	"net/url"
	"os"
	"path/filepath"
)

type ConfigTarantool struct {
	Address  string
	User     string
	Password string
}

type ConfigMattermost struct {
	UserName  string
	TeamName  string
	Token     string
	Channel   string
	ServerURL *url.URL
}

type Config struct {
	Tarantool  ConfigTarantool
	Mattermost ConfigMattermost
}

func LoadConfig() Config {
	var settings Config

	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	err = godotenv.Load(filepath.Join(pwd, "../.env"))
	if err != nil {
		log.Fatal("Не удалось загрузить данные из .env файла %v", err)
	}

	settings.Tarantool.User = os.Getenv("TT_USER")
	settings.Tarantool.Password = os.Getenv("TT_PASSWORD")
	settings.Tarantool.Address = os.Getenv("TT_ADDRESS")

	settings.Mattermost.TeamName = os.Getenv("MM_TEAM")
	settings.Mattermost.UserName = os.Getenv("MM_USERNAME")
	settings.Mattermost.Token = os.Getenv("MM_TOKEN")
	settings.Mattermost.Channel = os.Getenv("MM_CHANNEL")
	settings.Mattermost.ServerURL, _ = url.Parse(os.Getenv("MM_SERVER_URL"))

	return settings
}
