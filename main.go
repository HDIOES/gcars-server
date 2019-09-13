package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/HDIOES/gcars-server/core"
	"github.com/HDIOES/gcars-server/util"
	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
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
	session := core.Session{}
	http.HandleFunc("/receive", func(w http.ResponseWriter, r *http.Request) {
		upgrader.CheckOrigin = func(r *http.Request) bool {
			return true
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		session.AddCar(2000, 5000, conn)
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
