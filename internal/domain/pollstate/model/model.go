package model

import "time"

type Main struct {
	AccountID    int64
	LastMatchID  *int64
	LastPolledAt *time.Time
}
