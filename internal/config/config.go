package config

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const formatJSON = "json"

type Config struct {
	Server struct {
		Host        string `envconfig:"SERVER_HOST" default:":9000"`
		MetricsBind string `envconfig:"BIND_METRICS" default:":9090"`
		HealthHost  string `envconfig:"BIND_HEALTH" default:":9091"`
	}

	Service struct {
		LogLevel  string `envconfig:"LOGGER_LEVEL" default:"debug"`
		LogFormat string `envconfig:"LOGGER_FORMAT" default:"console"`
	}

	DB struct {
		Address  string `envconfig:"DB_ADDRESS"`
		Name     string `envconfig:"DB_NAME"`
		User     string `envconfig:"DB_USER"`
		Password string `envconfig:"DB_PASSWORD"`
		Port     string `envconfig:"DB_PORT"`
		MaxConn  int    `envconfig:"DB_MAX_CONN"`
	}
}

func Parse() (*Config, error) {
	var cfg = &Config{}
	err := envconfig.Process("", cfg)

	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (cfg Config) Logger() (logger zerolog.Logger) {
	level := zerolog.InfoLevel
	if newLevel, err := zerolog.ParseLevel(cfg.Service.LogLevel); err == nil {
		level = newLevel
	}

	var out io.Writer = os.Stdout
	if cfg.Service.LogFormat != formatJSON {
		out = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.StampMicro}
	}
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	return zerolog.New(out).Level(level).With().Caller().Timestamp().Logger()
}

func (cfg Config) PgPoolConfig() (*pgxpool.Config, error) {
	poolCfg, err := pgxpool.ParseConfig(fmt.Sprintf(
		"host=%s port=%d dbname=%s sslmode=disable user=%s password=%s pool_max_conns=%d",
		cfg.DB.Address, cfg.DB.Port, cfg.DB.Name, cfg.DB.User, cfg.DB.Password, cfg.DB.MaxConn,
	))
	if err != nil {
		return nil, err
	}

	return poolCfg, nil
}

// Для Google аутентификации
func SetupConfig() *oauth2.Config {
	conf := &oauth2.Config{
		RedirectURL:  "http://localhost:8080/callback",
		ClientID:     "21646045870-l985handt84rtn20d5m0htd9a5t6a0rv.apps.googleusercontent.com",
		ClientSecret: "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",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
	return conf
}
