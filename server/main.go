package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/caarlos0/env/v11"
	_ "github.com/joho/godotenv/autoload"
	"github.com/zainokta/client-server-multiplayer/server/config"
	"github.com/zainokta/client-server-multiplayer/server/game"
)

func startUDPServer(cfg config.Config) {
	addr := net.UDPAddr{Port: cfg.Port, IP: net.ParseIP("127.0.0.1")}
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	fmt.Printf("UDP Server listening on port %d\n", cfg.Port)

	gameState := game.New()

	ticker := time.NewTicker(time.Second / time.Duration(cfg.GameTickRate))

	go func() {
		for range ticker.C {
			gameState.Broadcast(conn)
		}
	}()

	for {
		buf := make([]byte, 1024)
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Println("Error reading:", err)
			continue
		}

		go gameState.HandleClient(conn, addr, buf[:n])
	}
}

func main() {
	cfg, err := env.ParseAs[config.Config]()
	if err != nil {
		fmt.Printf("%+v\n", err)
	}

	startUDPServer(cfg)
}
