package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yerlan/dota2/internal/config"
)

type runner interface {
	Run(ctx context.Context)
}

type App struct {
	pgpool *pgxpool.Pool

	httpServer *http.Server

	tgListener runner
	poller     runner

	ctx       context.Context
	ctxCancel context.CancelFunc

	exitCode int
}

func (a *App) Init() {
	var err error

	a.ctx, a.ctxCancel = context.WithCancel(context.Background())

	initLogger(config.Conf.Debug, config.Conf.LogLevel)

	a.pgpool, err = initPgPool(config.Conf.PgDsn)
	errCheck(err, "pgpool init")

	runMigrations()
	slog.Info("PG-migrations have been successfully applied")

	a.wire()

	a.httpServer = &http.Server{
		Addr:              ":" + config.Conf.HTTPPort,
		Handler:           newHealthMux(),
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       time.Minute,
	}
}

func (a *App) PreStartHook() {
	slog.Info("PreStartHook")
}

func (a *App) Start() {
	slog.Info("Starting")

	if a.tgListener != nil {
		go a.tgListener.Run(a.ctx)
		slog.Info("telegram listener started")
	}

	if a.poller != nil {
		go a.poller.Run(a.ctx)
		slog.Info("poller started")
	}

	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server", "error", err)
		}
	}()
	slog.Info("http server started " + a.httpServer.Addr)
}

func (a *App) Listen() {
	signalCtx, signalCtxCancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer signalCtxCancel()

	<-signalCtx.Done()
}

func (a *App) Stop() {
	slog.Info("Shutting down...")

	a.ctxCancel()

	ctx, ctxCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer ctxCancel()

	if err := a.httpServer.Shutdown(ctx); err != nil {
		slog.Error("http-server shutdown error", "error", err)
		a.exitCode = 1
	}
}

func (a *App) WaitJobs() {
	slog.Info("waiting jobs")
	time.Sleep(500 * time.Millisecond)
}

func (a *App) Exit() {
	slog.Info("Exit")

	if a.pgpool != nil {
		a.pgpool.Close()
	}

	os.Exit(a.exitCode)
}

func newHealthMux() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	return mux
}

func errCheck(err error, msg string) {
	if err != nil {
		if msg != "" {
			err = fmt.Errorf("%s: %w", msg, err)
		}
		slog.Error(err.Error())
		os.Exit(1)
	}
}
