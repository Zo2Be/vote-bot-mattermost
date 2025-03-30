package bot

import (
	"github.com/mattermost/mattermost-server/v6/model"
	"sync"
	"vote-bot/internal/db"
)

type Bot struct {
	client    *model.Client4
	botUser   *model.User
	Channel   *model.Channel
	dbConn    *db.ConnTarantool
	pollsLock sync.RWMutex
}
