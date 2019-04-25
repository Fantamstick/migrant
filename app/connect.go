package app

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

// MustConnect will connect to the specified database or log a fatal.
func MustConnect(config DatabaseConfig) *sql.DB {
	if config.PortForward {
		initPortforwarding(config)
	}

	con, err := sql.Open(config.Driver, config.Uri)

	if err != nil {
		log.Fatal(err)
	}

	return con
}

// initialize port forwarding if required
func initPortforwarding(config DatabaseConfig) {
	t, err := NewTunnel(config.TunnelConfig)

	if err != nil {
		log.Fatal(err)
	}

	ready := make(chan bool)
	go t.Start(ready)
	<-ready
}
