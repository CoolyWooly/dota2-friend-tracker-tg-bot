package app

import (
	"github.com/yerlan/dota2/internal/config"

	accountdb "github.com/yerlan/dota2/internal/domain/account/repo/db"
	accountsvc "github.com/yerlan/dota2/internal/domain/account/service"
	notifdb "github.com/yerlan/dota2/internal/domain/notification/repo/db"
	notifsvc "github.com/yerlan/dota2/internal/domain/notification/service"
	pollstatedb "github.com/yerlan/dota2/internal/domain/pollstate/repo/db"
	pollstatesvc "github.com/yerlan/dota2/internal/domain/pollstate/service"

	tglistener "github.com/yerlan/dota2/internal/handler/telegram"

	"github.com/yerlan/dota2/internal/service/heroname"
	"github.com/yerlan/dota2/internal/service/opendota"
	"github.com/yerlan/dota2/internal/service/poller"
	tgsvc "github.com/yerlan/dota2/internal/service/telegram"

	botuc "github.com/yerlan/dota2/internal/usecase/bot"
	matchuc "github.com/yerlan/dota2/internal/usecase/match"
)

func (a *App) wire() {
	// repo
	accountRepo := accountdb.New(a.pgpool)
	pollStateRepo := pollstatedb.New(a.pgpool)
	notifRepo := notifdb.New(a.pgpool)

	// domain services
	accountSvc := accountsvc.New(accountRepo)
	pollStateSvc := pollstatesvc.New(pollStateRepo)
	notifSvc := notifsvc.New(notifRepo)

	// external services
	openDota := opendota.New(config.Conf.OpenDotaAPIKey)
	heroNamer := heroname.New(openDota, config.Conf.HeroNameTTL)

	tgBot, err := tgsvc.New(config.Conf.TgToken, config.Conf.TgDebug)
	errCheck(err, "telegram bot init")

	// usecases
	matchUC := matchuc.New(config.Conf.TgOwnerID, accountSvc, notifSvc, heroNamer, openDota, tgBot)
	botUC := botuc.New(config.Conf.TgOwnerID, config.Conf.PollInterval, accountSvc, openDota, tgBot)

	// transport + background
	a.tgListener = tglistener.New(tgBot, botUC)
	a.poller = poller.New(config.Conf.PollInterval, accountSvc, pollStateSvc, openDota, matchUC)
}
