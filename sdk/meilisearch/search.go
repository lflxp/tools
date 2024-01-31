package meilisearch

import (
	"log/slog"

	"github.com/meilisearch/meilisearch-go"
)

var (
	SearchCli    *meilisearch.Client
	Host, Apikey string
)

func InitMeili() {
	slog.Debug("Meilisearch", "host", Host, "apikey", Apikey)
	SearchCli = meilisearch.NewClient(meilisearch.ClientConfig{
		Host:   Host,
		APIKey: Apikey,
	})
}
