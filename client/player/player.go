package player

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"math/rand/v2"
	"net"
	"sync"
	"time"
)

const (
	InterpolationAlpha float32 = 0.1
)

var sequenceNumber uint32 = 0

type Player struct {
	ID        int32
	X         float32
	Y         float32
	Timestamp int64
	Sequence  uint32
}

var LastPositions sync.Map

func New() Player {
	return Player{
		ID: int32(rand.IntN(8)),
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

func predictPosition(last Player, deltaTime float32) Player {
	velocityX := (last.X - last.X) / deltaTime
	velocityY := (last.Y - last.Y) / deltaTime
	last.X += velocityX * deltaTime
	last.Y += velocityY * deltaTime
	return last
}

func interpolatePosition(last, new Player, alpha float32) Player {
	return Player{
		ID:        last.ID,
		X:         last.X + (new.X-last.X)*alpha,
		Y:         last.Y + (new.Y-last.Y)*alpha,
		Timestamp: new.Timestamp,
		Sequence:  new.Sequence,
	}
}

func receiveUpdate(buf []byte, n int) error {
	now := time.Now().UnixMilli()

	updatedPlayer, err := deserializePlayer(buf[:n])
	if err != nil {
		log.Println("Failed to decode: ", err)
		return err
	}

	val, exists := LastPositions.Load(updatedPlayer.ID)
	if exists {
		lastPlayerPosition := val.(Player)

		// calculate in seconds
		deltaTime := float32(now-lastPlayerPosition.Timestamp) / 1000.0

		if deltaTime > 1.0 {
			predicted := predictPosition(lastPlayerPosition, deltaTime)
			fmt.Printf("[Client] Player %d Predicted: X=%.2f, Y=%.2f (delta time=%.2fs)\n",
				predicted.ID, predicted.X, predicted.Y, deltaTime)
		}

		interpolated := interpolatePosition(lastPlayerPosition, updatedPlayer, InterpolationAlpha)

		fmt.Printf("[Client] Player %d Interpolated: X=%.2f, Y=%.2f\n", interpolated.ID, interpolated.X, interpolated.Y)
	} else {
		fmt.Printf("[Client] Player %d First Update: X=%.2f, Y=%.2f\n", updatedPlayer.ID, updatedPlayer.X, updatedPlayer.Y)
	}

	updatedPlayer.Timestamp = now
	LastPositions.Store(updatedPlayer.ID, updatedPlayer)

	return nil
}

func getPlayerUpdate(conn *net.UDPConn) {
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

func SendPlayerUpdate(conn *net.UDPConn) {
	gamePlayer := New()

	go getPlayerUpdate(conn)

	for i := 0; i < 10; i++ {
		sequenceNumber++
		gamePlayer.X = rand.Float32() * 100
		gamePlayer.Y = rand.Float32() * 100
		gamePlayer.Timestamp = time.Now().UnixMilli()
		gamePlayer.Sequence = sequenceNumber

		data, err := serializePlayer(gamePlayer)
		if err != nil {
			log.Fatal(err)
		}

		_, err = conn.Write(data)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Sent Player %d: X=%.2f, Y=%.2f\n", gamePlayer.ID, gamePlayer.X, gamePlayer.Y)
		time.Sleep(time.Second)
	}
}
