package game

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/zainokta/client-server-multiplayer/server/player"
)

type GameState struct {
	Players sync.Map
	Clients sync.Map
}

func New() GameState {
	return GameState{}
}

func (g *GameState) HandleClient(conn *net.UDPConn, addr *net.UDPAddr, data []byte) {
	player, err := player.DeserializePlayer(data)
	if err != nil {
		log.Println("Failed to decode player data:", err)
		return
	}

	g.Players.Store(player.ID, player)
	g.Clients.Store(player.ID, addr)

	fmt.Printf("Player %d updated: X=%d, Y=%d (from %s)\n", player.ID, player.X, player.Y, addr)
}

func (g *GameState) Broadcast(conn *net.UDPConn) {
	g.Players.Range(func(key, value interface{}) bool {
		p := value.(player.Player)

		data, err := player.SerializePlayer(p)
		if err != nil {
			log.Println("Error serializing player:", err)
			return true
		}

		g.Clients.Range(func(_, addr interface{}) bool {
			udpAddr := addr.(*net.UDPAddr)
			_, err := conn.WriteToUDP(data, udpAddr)
			if err != nil {
				log.Println("Error broadcasting:", err)
			}
			return true
		})

		return true
	})
}
