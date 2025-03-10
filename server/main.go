package main

import (
	"fmt"
	"log"
	"net"

	"github.com/caarlos0/env/v11"
	_ "github.com/joho/godotenv/autoload"
	"github.com/zainokta/client-server-multiplayer/server/config"
	"github.com/zainokta/client-server-multiplayer/server/player"
)

func startUDPServer(cfg config.Config) {
	addr := net.UDPAddr{Port: cfg.Port, IP: net.ParseIP("127.0.0.1")}
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	fmt.Printf("UDP Server listening on port %d\n", cfg.Port)

	for {
		buf := make([]byte, 1024)
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Println("Error reading:", err)
			continue
		}

		player, err := player.DeserializePlayer(buf[:n])
		if err != nil {
			log.Println("Error decoding:", err)
			continue
		}

		fmt.Printf("Received from %s - Player %d: X=%d, Y=%d\n", addr, player.ID, player.X, player.Y)
	}
}

func main() {
	cfg, err := env.ParseAs[config.Config]()
	if err != nil {
		fmt.Printf("%+v\n", err)
	}

	startUDPServer(cfg)
}
