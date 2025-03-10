package player

import (
	"bytes"
	"encoding/binary"
)

type Player struct {
	ID int32
	X  int32
	Y  int32
}

var playerId int32 = 0

func New() Player {
	playerId++
	return Player{
		ID: playerId,
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
