package model

import "time"

type Main struct {
	AccountID int64
	MatchID   int64
	SentAt    time.Time
}
