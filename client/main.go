package main

import (
	"fmt"
	"log"
	"net"

	"github.com/caarlos0/env/v11"
	"github.com/zainokta/client-server-multiplayer/client/config"
	"github.com/zainokta/client-server-multiplayer/client/player"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	cfg, err := env.ParseAs[config.Config]()
	if err != nil {
		fmt.Printf("%+v\n", err)
	}

	addr := net.UDPAddr{Port: cfg.Port, IP: net.ParseIP("127.0.0.1")}
	conn, err := net.DialUDP("udp", nil, &addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	player.SendPlayerUpdate(conn)
}
