package telegram

import (
	"fmt"
	"html"
	"strconv"
	"strings"
	"time"

	"github.com/yerlan/dota2/internal/constant"
	"github.com/yerlan/dota2/internal/service/opendota"
)

func FormatMatchCard(m *opendota.RecentMatch, heroName, ownerLabel string) string {
	resultEmoji := "🟢 Победа"
	if !m.IsWin() {
		resultEmoji = "🔴 Поражение"
	}

	header := resultEmoji
	if ownerLabel != "" {
		header = fmt.Sprintf("%s — %s", resultEmoji, html.EscapeString(ownerLabel))
	}

	var b strings.Builder
	fmt.Fprintf(&b, "<b>%s</b>\n", header)
	fmt.Fprintf(&b, "Герой: <b>%s</b>\n", html.EscapeString(heroName))
	fmt.Fprintf(&b, "KDA: %d/%d/%d\n", m.Kills, m.Deaths, m.Assists)
	fmt.Fprintf(&b, "GPM/XPM: %d / %d\n", m.GoldPerMin, m.XPPerMin)
	fmt.Fprintf(&b, "LH/DN: %d / %d\n", m.LastHits, m.Denies)
	fmt.Fprintf(&b, "Длительность: %s\n", formatDuration(m.Duration))
	fmt.Fprintf(&b, "Режим: %s | Лобби: %s\n", gameModeName(m.GameMode), lobbyTypeName(m.LobbyType))
	fmt.Fprintf(&b, `<a href="%s%d">Открыть матч</a>`, constant.DotaMatchURL, m.MatchID)
	return b.String()
}

func formatDuration(seconds int) string {
	d := time.Duration(seconds) * time.Second
	m := int(d.Minutes())
	s := int(d.Seconds()) - m*60
	return fmt.Sprintf("%d:%02d", m, s)
}

var gameModes = map[int]string{
	0: "Unknown", 1: "All Pick", 2: "Captains Mode", 3: "Random Draft",
	4: "Single Draft", 5: "All Random", 6: "Intro", 7: "Diretide",
	8: "Reverse Captains Mode", 9: "The Greeviling", 10: "Tutorial", 11: "Mid Only",
	12: "Least Played", 13: "Limited Heroes", 14: "Compendium Matchmaking",
	15: "Custom", 16: "Captains Draft", 17: "Balanced Draft", 18: "Ability Draft",
	19: "Event", 20: "All Random Death Match", 21: "1v1 Solo Mid",
	22: "Ranked All Pick", 23: "Turbo", 24: "Mutation",
}

func gameModeName(id int) string {
	if n, ok := gameModes[id]; ok {
		return n
	}
	return "mode " + strconv.Itoa(id)
}

var lobbyTypes = map[int]string{
	0: "Normal", 1: "Practice", 2: "Tournament", 3: "Tutorial",
	4: "Co-op Bots", 5: "Team", 6: "Solo Queue", 7: "Ranked",
	8: "1v1 Mid", 9: "Battle Cup",
}

func lobbyTypeName(id int) string {
	if n, ok := lobbyTypes[id]; ok {
		return n
	}
	return "lobby " + strconv.Itoa(id)
}
