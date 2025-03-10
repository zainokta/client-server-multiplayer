package player

import (
	"bytes"
	"encoding/binary"
	"log"
	"math/rand/v2"
	"net"
	"sync"
	"time"
)

const (
	SmoothingFactor = 0.2
	MaxRewindTime   = 300 * time.Millisecond
	MinTimeDiff     = 0.001
)

type Player struct {
	ID        int32
	X         float32
	Y         float32
	Timestamp int64
	Sequence  uint32
}

var Players sync.Map

func New(x float32, y float32) Player {
	return Player{
		ID: int32(rand.IntN(8)),
		X:  x,
		Y:  y,
	}
}

func serializePlayer(player Player) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, player)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func deserializePlayer(data []byte) (Player, error) {
	var player Player
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.LittleEndian, &player)
	return player, err
}

func interpolatePlayer(last, current Player) Player {
	current.X = last.X + (current.X-last.X)*SmoothingFactor
	current.Y = last.Y + (current.Y-last.Y)*SmoothingFactor
	return current
}

func reconcilePlayerPosition(gamePlayer Player) {
	lastVal, exists := Players.Load(gamePlayer.ID)
	if exists {
		clientPlayer := lastVal.(Player)
		if gamePlayer.Sequence <= clientPlayer.Sequence {
			return
		}

		if time.Now().UnixMilli()-gamePlayer.Timestamp > int64(MaxRewindTime) {
			clientPlayer.X = gamePlayer.X
			clientPlayer.Y = gamePlayer.Y
		} else {
			clientPlayer = interpolatePlayer(clientPlayer, gamePlayer)
		}

		clientPlayer.Sequence = gamePlayer.Sequence
		Players.Store(clientPlayer.ID, clientPlayer)
	}
}

func receiveUpdate(buf []byte, n int) error {
	now := time.Now().UnixMilli()

	updatedPlayer, err := deserializePlayer(buf[:n])
	if err != nil {
		log.Println("Failed to decode: ", err)
		return err
	}

	reconcilePlayerPosition(updatedPlayer)

	updatedPlayer.Timestamp = now
	Players.Store(updatedPlayer.ID, updatedPlayer)

	return nil
}

func PredictPosition(p Player) Player {
	lastVal, exists := Players.Load(p.ID)
	if exists {
		lastPlayer := lastVal.(Player)
		timeDiff := float32(p.Timestamp-lastPlayer.Timestamp) / 1000.0
		if timeDiff < MinTimeDiff {
			return lastPlayer
		}

		velX := (p.X - lastPlayer.X) / timeDiff
		velY := (p.Y - lastPlayer.Y) / timeDiff

		p.X += velX * timeDiff
		p.Y += velY * timeDiff
	}

	return p
}

func GetPlayerUpdate(conn *net.UDPConn) {
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Println("Error receiving: ", err)
			continue
		}

		receiveUpdate(buf, n)
	}
}

func SendPlayerUpdate(conn *net.UDPConn, gamePlayer Player) {
	data, err := serializePlayer(gamePlayer)
	if err != nil {
		log.Fatal(err)
	}

	_, err = conn.Write(data)
	if err != nil {
		log.Fatal(err)
	}
}
