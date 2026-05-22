package opendota

type Player struct {
	Profile struct {
		AccountID   int64  `json:"account_id"`
		Personaname string `json:"personaname"`
	} `json:"profile"`
	RankTier int `json:"rank_tier"`
}

type RecentMatch struct {
	MatchID    int64 `json:"match_id"`
	HeroID     int   `json:"hero_id"`
	StartTime  int64 `json:"start_time"`
	Duration   int   `json:"duration"`
	GameMode   int   `json:"game_mode"`
	LobbyType  int   `json:"lobby_type"`
	PlayerSlot int   `json:"player_slot"`
	RadiantWin bool  `json:"radiant_win"`
	Kills      int   `json:"kills"`
	Deaths     int   `json:"deaths"`
	Assists    int   `json:"assists"`
	LastHits   int   `json:"last_hits"`
	Denies     int   `json:"denies"`
	GoldPerMin int   `json:"gold_per_min"`
	XPPerMin   int   `json:"xp_per_min"`
}

func (m *RecentMatch) IsRadiant() bool { return m.PlayerSlot < 128 }
func (m *RecentMatch) IsWin() bool     { return m.IsRadiant() == m.RadiantWin }

type HeroStat struct {
	ID            int    `json:"id"`
	LocalizedName string `json:"localized_name"`
}
