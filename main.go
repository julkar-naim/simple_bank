package main

import (
	"database/sql"
	"github.com/julkar-naim/simple-bank/api"
	db "github.com/julkar-naim/simple-bank/db/sqlc"
	"github.com/julkar-naim/simple-bank/util"
	_ "github.com/lib/pq"
	"log"
)

func main() {
	config, err := util.LoadConfig()
	if err != nil {
		log.Fatal("cant read all the config", err)
	}
	util.NewConfig(config)

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to database", err)
	}
	store := db.NewStore(conn)
	server := api.NewServer(store)

	err = server.Start(config.Address)
	if err != nil {
		log.Fatal("cannot start server", err)
	}
}
