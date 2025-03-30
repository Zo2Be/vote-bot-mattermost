package poll

import "time"

type Poll struct {
	ID        string
	Creator   string
	Question  string
	Options   []string
	Votes     map[string]int
	Active    bool
	CreatedAt time.Time
	ClosedAt  *time.Time
}
