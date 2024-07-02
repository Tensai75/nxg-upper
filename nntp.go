package main

import (
	"fmt"
	"time"

	"github.com/Tensai75/nntpPool"
)

var (
	pool            nntpPool.ConnectionPool
	headerCheckPool nntpPool.ConnectionPool
)

func initNntpPool() {
	var err error

	uploadConnections := conf.Connections

	go func() {
		for {
			select {
			case v := <-nntpPool.LogChan:
				Log.Debug(fmt.Sprintf("NNTPPool: %v", v))
			case w := <-nntpPool.WarnChan:
				Log.Warn(fmt.Sprintf("NNTPPool Error: %v", w.Error()))
			}
		}
	}()

	if conf.HeaderCheck {
		uploadConnections = conf.Connections - conf.HeaderCheckConns
		if uploadConnections < 1 {
			checkForFatalErr(fmt.Errorf("HeaderCheckConns must be smaller than the number of connections"))
		}
		if headerCheckPool, err = nntpPool.New(&nntpPool.Config{
			Host:                  conf.Host,
			Port:                  conf.Port,
			SSL:                   conf.SSL,
			SkipSSLCheck:          true,
			User:                  conf.NntpUser,
			Pass:                  conf.NntpPass,
			ConnWaitTime:          time.Duration(conf.ConnWaitTime) * time.Second,
			MaxConns:              conf.HeaderCheckConns,
			IdleTimeout:           30 * time.Second,
			MaxConnErrors:         conf.Retries,
			MaxTooManyConnsErrors: 3,
		}, 0); err != nil {
			checkForFatalErr(err)
		}
	}

	if pool, err = nntpPool.New(&nntpPool.Config{
		Host:                  conf.Host,
		Port:                  conf.Port,
		SSL:                   conf.SSL,
		SkipSSLCheck:          true,
		User:                  conf.NntpUser,
		Pass:                  conf.NntpPass,
		ConnWaitTime:          time.Duration(conf.ConnWaitTime) * time.Second,
		MaxConns:              uploadConnections,
		IdleTimeout:           30 * time.Second,
		MaxConnErrors:         conf.Retries,
		MaxTooManyConnsErrors: 3,
	}, 0); err != nil {
		checkForFatalErr(err)
	}
}
