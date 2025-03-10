package player

import (
	"bytes"
	"encoding/binary"
)

type Player struct {
	ID        int32
	X         float32
	Y         float32
	Timestamp int64
	Sequence  uint32
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
	var gamePlayer Player
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.LittleEndian, &gamePlayer)
	return gamePlayer, err
}
