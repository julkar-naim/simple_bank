package main

import (
	"database/sql"
	"github.com/julkar-naim/simple-bank/api"
	db "github.com/julkar-naim/simple-bank/db/sqlc"
	_ "github.com/lib/pq"
	"log"
)

const (
	dbDriver = "postgres"
	dbUri    = "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"
	address  = "0.0.0.0:8080"
)

func main() {
	conn, err := sql.Open(dbDriver, dbUri)
	if err != nil {
		log.Fatal("cannot connect to database", err)
	}
	store := db.NewStore(conn)
	server := api.NewServer(store)

	err = server.Start(address)
	if err != nil {
		log.Fatal("cannot start server", err)
	}
}
