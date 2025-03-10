package main

import (
	"fmt"
	"log"
	"math/rand/v2"
	"net"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/zainokta/client-server-multiplayer/client/config"
	"github.com/zainokta/client-server-multiplayer/client/player"

	_ "github.com/joho/godotenv/autoload"
)

func sendUDPPlayerUpdate(cfg config.Config) {
	addr := net.UDPAddr{Port: cfg.Port, IP: net.ParseIP("127.0.0.1")}
	conn, err := net.DialUDP("udp", nil, &addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	p := player.New()

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				log.Println("Error receiving: ", err)
				continue
			}

			updatedPlayer, err := player.DeserializePlayer(buf[:n])
			if err != nil {
				log.Println("Failed to decode: ", err)
				continue
			}

			fmt.Printf("[Client] Player %d: X=%d, Y=%d\n", updatedPlayer.ID, updatedPlayer.X, updatedPlayer.Y)
		}
	}()

	for i := 0; i < 10; i++ {
		p.X = rand.Int32() * 100
		p.Y = rand.Int32() * 100

		data, err := player.SerializePlayer(p)
		if err != nil {
			log.Fatal(err)
		}

		_, err = conn.Write(data)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Sent Player %d: X=%d, Y=%d\n", p.ID, p.X, p.Y)
		time.Sleep(time.Second)
	}
}

func main() {
	cfg, err := env.ParseAs[config.Config]()
	if err != nil {
		fmt.Printf("%+v\n", err)
	}

	sendUDPPlayerUpdate(cfg)
}
