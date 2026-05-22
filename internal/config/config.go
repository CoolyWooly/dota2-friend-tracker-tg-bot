package config

import (
	"time"

	"github.com/caarlos0/env/v9"
	_ "github.com/joho/godotenv/autoload"
)

var Conf = struct {
	Namespace string `env:"NAMESPACE" envDefault:"dota2-bot"`
	Debug     bool   `env:"DEBUG" envDefault:"false"`
	LogLevel  string `env:"LOG_LEVEL" envDefault:"info"`

	HTTPPort string `env:"HTTP_PORT" envDefault:"8080"`

	PgDsn string `env:"PG_DSN"`

	TgToken   string `env:"TG_TOKEN,required"`
	TgOwnerID int64  `env:"OWNER_TG_ID,required"`
	TgDebug   bool   `env:"TG_DEBUG" envDefault:"false"`

	OpenDotaAPIKey string `env:"OPENDOTA_API_KEY"`

	PollInterval time.Duration `env:"POLL_INTERVAL" envDefault:"10m"`
	HeroNameTTL  time.Duration `env:"HERO_NAME_TTL" envDefault:"24h"`
}{}

func init() {
	if err := env.Parse(&Conf); err != nil {
		panic(err)
	}
}
