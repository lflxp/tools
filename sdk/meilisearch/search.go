package meilisearch

import (
	log "github.com/go-eden/slf4go"
	"github.com/meilisearch/meilisearch-go"
)

var (
	SearchCli    *meilisearch.Client
	Host, Apikey string
)

func InitMeili() {
	log.Debugf("Meilisearch host %s apikey %s", Host, Apikey)
	SearchCli = meilisearch.NewClient(meilisearch.ClientConfig{
		Host:   Host,
		APIKey: Apikey,
	})
}
