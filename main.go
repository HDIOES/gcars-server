package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/HDIOES/gcars-server/game"
	"github.com/HDIOES/gcars-server/util"
	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/tkanos/gonfig"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  8192,
	WriteBufferSize: 8192,
}

func main() {
	log.Println("Start server...")
	configuration := util.Configuration{}
	gonfigErr := gonfig.GetConf("configuration.json", &configuration)
	if gonfigErr != nil {
		panic(gonfigErr)
	}
	if len(os.Args) > 1 && strings.Compare(os.Args[1], "dbmode") == 0 {
		db, err := sql.Open("postgres", configuration.DatabaseURL)
		if err != nil {
			panic(err)
		}
		db.SetMaxIdleConns(configuration.MaxIdleConnections)
		db.SetMaxOpenConns(configuration.MaxOpenConnections)
		timeout := strconv.Itoa(configuration.ConnectionTimeout) + "s"
		timeoutDuration, durationErr := time.ParseDuration(timeout)
		if durationErr != nil {
			log.Println("Error parsing of timeout parameter")
			panic(durationErr)
		} else {
			db.SetConnMaxLifetime(timeoutDuration)
		}

		migrations := &migrate.FileMigrationSource{
			Dir: "migrations",
		}

		if n, err := migrate.Exec(db, "postgres", migrations, migrate.Up); err == nil {
			log.Printf("Applied %d migrations!\n", n)
		} else {
			log.Panic(err)
		}
	}
	var serverInstance = game.CreateServerInstance()
	serverInstance.CreateSession()
	http.HandleFunc("/receive", func(w http.ResponseWriter, r *http.Request) {
		upgrader.CheckOrigin = func(r *http.Request) bool {
			return true
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		serverInstance.Sessions[0].CreatePlayer(conn)
		if err != nil {
			conn.Close()
			return
		}
		log.Println("Connection added")
	})

	log.Println("This server is running!")
	if err := http.ListenAndServe(":"+strconv.Itoa(configuration.Port), nil); err != nil {
		log.Panic(err)
	}
}
