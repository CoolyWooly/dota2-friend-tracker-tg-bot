package constant

import "time"

const (
	ServiceName = "Dota2"

	OpenDotaBaseURL = "https://api.opendota.com/api"
	DotaMatchURL    = "https://www.opendota.com/matches/"

	HTTPClientTimeout = 15 * time.Second
	OpenDotaRPS       = 1
)
