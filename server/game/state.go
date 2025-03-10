package game

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/zainokta/client-server-multiplayer/server/player"
)

const (
	disconnectTimer = 5 * time.Second
)

type GameState struct {
	Players         sync.Map
	Clients         sync.Map
	SequenceNumbers sync.Map
}

func New() GameState {
	return GameState{}
}

func (g *GameState) HandleClient(conn *net.UDPConn, addr *net.UDPAddr, data []byte) {
	gamePlayer, err := player.DeserializePlayer(data)
	if err != nil {
		log.Println("Failed to decode player data:", err)
		return
	}

	if lastSeq, exists := g.SequenceNumbers.Load(gamePlayer.ID); exists {
		if gamePlayer.Sequence <= lastSeq.(uint32) {
			fmt.Printf("[Server] Ignoring outdated packet from Player %d\n", gamePlayer.ID)
			return
		}
	}

	g.SequenceNumbers.Store(gamePlayer.ID, gamePlayer.Sequence)

	g.Players.Store(gamePlayer.ID, gamePlayer)
	g.Clients.Store(gamePlayer.ID, addr)

	fmt.Printf("Player %d updated: X=%.2f, Y=%.2f (from %s)\n", gamePlayer.ID, gamePlayer.X, gamePlayer.Y, addr)
}

func (g *GameState) Broadcast(conn *net.UDPConn) {
	g.Players.Range(func(key, value interface{}) bool {
		p := value.(player.Player)

		data, err := player.SerializePlayer(p)
		if err != nil {
			log.Println("Error serializing player:", err)
			return true
		}

		g.Clients.Range(func(key, addr interface{}) bool {
			udpAddr := addr.(*net.UDPAddr)
			_, err := conn.WriteToUDP(data, udpAddr)
			if err != nil {
				log.Println("Error broadcasting:", err)
				g.Clients.Delete(key)
			}
			return true
		})

		return true
	})
}

func (g *GameState) MonitorDisconnections() {
	ticker := time.NewTicker(disconnectTimer)
	for range ticker.C {
		now := time.Now().UnixMilli()
		g.Players.Range(func(key, value interface{}) bool {
			gamePlayer := value.(player.Player)
			if now-gamePlayer.Timestamp > disconnectTimer.Milliseconds() {
				fmt.Printf("[Server] Player %d disconnected.\n", gamePlayer.ID)
				g.Players.Delete(key)
				g.SequenceNumbers.Delete(key)
				g.Clients.Delete(key)
			}
			return true
		})
	}
}
