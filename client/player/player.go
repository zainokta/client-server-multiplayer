package player

import (
	"bytes"
	"encoding/binary"
	"math/rand/v2"
)

type Player struct {
	ID int32
	X  int32
	Y  int32
}

func New() Player {
	return Player{
		ID: int32(rand.IntN(8)),
	}
}

func SerializePlayer(player Player) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, player)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func DeserializePlayer(data []byte) (Player, error) {
	var player Player
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.LittleEndian, &player)
	return player, err
}
