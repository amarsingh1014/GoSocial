package main

import (
    "fmt"
    "log"
    "social/internal/db"
    "social/internal/env"
    "social/internal/store"
)

const version = "0.0.1"

type config struct {
    addr   string
    dbconn dbconfig
    env    string
}

type dbconfig struct {
    addr         string
    maxOpenConns int
    maxIdleConns int
    maxIdletime  string
}

func main() {

    cfg := config{
        addr: env.GetString("ADDR", ":8080"),
        dbconn: dbconfig{
            addr:         env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost:5433/social_network?sslmode=disable"),
            maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 25),
            maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 25),
            maxIdletime:  env.GetString("DB_MAX_OPEN_CONNS_LIFETIME", "15m"),
        },
        env: env.GetString("ENV", "development"),
    }

    dbConn, err := db.New(
        cfg.dbconn.addr,
        cfg.dbconn.maxOpenConns,
        cfg.dbconn.maxIdleConns,
        cfg.dbconn.maxIdletime,
    )

    if err != nil {
        log.Panic(err)
    }

    defer dbConn.Close()

    fmt.Println("Database connection established")

    store := store.NewStorage(dbConn)

    // Seed the database
    if err := db.Seed(store); err != nil {
        log.Fatalf("Failed to seed database: %v", err)
    }

    fmt.Println("Database seeding completed successfully")
}