package main

import (
	"fmt"
	"log"
	"social/internal/db"
	"social/internal/env"
	"social/internal/store"

)

const version = "0.0.1"

func main() {

	cfg := config{
		addr: env.GetString("ADDR", ":8080"),
		dbconn: dbconfig{
			addr:         env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost:5433/social_network?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 25),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 25),
			maxIdletime:  env.GetString("DB_MAX_OPEN_CONNS_LIFETIME", "15m"),
		},
		env : env.GetString("ENV", "development"),
	}

	db, err := db.New(
		cfg.dbconn.addr,
		cfg.dbconn.maxOpenConns,
		cfg.dbconn.maxIdleConns,
		cfg.dbconn.maxIdletime,
	)

	if err != nil {
		log.Panic(err)
	}

	defer db.Close()

	fmt.Println("Database connection established")

	store := store.NewStorage(db)

	app := &application{
		config: cfg,
		store:  store,
	}

	mux := app.mount()

	log.Fatal(app.run(mux))

	fmt.Println("Server is running on port 8080")

}
