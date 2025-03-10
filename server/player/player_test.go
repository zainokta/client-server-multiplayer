package player

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPlayerSerialization(t *testing.T) {
	players := []Player{
		{
			ID:        1,
			X:         123.45,
			Y:         678.90,
			Timestamp: time.Now().UnixMilli(),
			Sequence:  42,
		},
		{
			ID:        2,
			X:         -50.5,
			Y:         0,
			Timestamp: 1647366824123,
			Sequence:  99999,
		},
		{
			ID:        -5,
			X:         0,
			Y:         -123.456,
			Timestamp: 0,
			Sequence:  0,
		},
	}

	for _, p := range players {
		data, err := SerializePlayer(p)
		assert.NoError(t, err)
		assert.NotEmpty(t, data)

		deserializedPlayer, err := DeserializePlayer(data)
		assert.NoError(t, err)

		assert.Equal(t, p.ID, deserializedPlayer.ID)
		assert.Equal(t, p.X, deserializedPlayer.X)
		assert.Equal(t, p.Y, deserializedPlayer.Y)
		assert.Equal(t, p.Timestamp, deserializedPlayer.Timestamp)
		assert.Equal(t, p.Sequence, deserializedPlayer.Sequence)
	}
}

func TestDeserializeInvalidData(t *testing.T) {
	_, err := DeserializePlayer([]byte{})
	assert.Error(t, err)

	_, err = DeserializePlayer([]byte{1, 2, 3})
	assert.Error(t, err)

	corruptData := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	_, err = DeserializePlayer(corruptData)
	assert.Error(t, err)
}

func TestRoundTripConsistency(t *testing.T) {
	original := Player{
		ID:        123,
		X:         456.78,
		Y:         -90.12,
		Timestamp: time.Now().UnixMilli(),
		Sequence:  34567,
	}

	data1, err := SerializePlayer(original)
	assert.NoError(t, err)

	player1, err := DeserializePlayer(data1)
	assert.NoError(t, err)

	data2, err := SerializePlayer(player1)
	assert.NoError(t, err)

	player2, err := DeserializePlayer(data2)
	assert.NoError(t, err)

	assert.Equal(t, original, player1, "First deserialization should match original")
	assert.Equal(t, player1, player2, "Second deserialization should match first")
	assert.True(t, bytes.Equal(data1, data2), "Serialized data should be identical across cycles")
}
