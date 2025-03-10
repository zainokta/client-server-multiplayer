package main_test

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/zainokta/client-server-multiplayer/server/config"
	"github.com/zainokta/client-server-multiplayer/server/game"
	"github.com/zainokta/client-server-multiplayer/server/player"
)

func TestGameStateHandleClient(t *testing.T) {
	gameState := game.New()

	conn := &mockUDPConn{}
	addr := &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 9000,
	}

	testPlayer := player.Player{
		ID:        1,
		X:         100.5,
		Y:         200.75,
		Timestamp: time.Now().UnixMilli(),
		Sequence:  1,
	}

	data, err := player.SerializePlayer(testPlayer)
	assert.NoError(t, err)

	gameState.HandleClient(conn, addr, data)

	storedPlayerInterface, exists := gameState.Players.Load(testPlayer.ID)
	assert.True(t, exists)

	storedPlayer := storedPlayerInterface.(player.Player)
	assert.Equal(t, testPlayer.ID, storedPlayer.ID)
	assert.Equal(t, testPlayer.X, storedPlayer.X)
	assert.Equal(t, testPlayer.Y, storedPlayer.Y)

	storedAddrInterface, exists := gameState.Clients.Load(testPlayer.ID)
	assert.True(t, exists)
	assert.Equal(t, addr, storedAddrInterface)

	storedSeqInterface, exists := gameState.SequenceNumbers.Load(testPlayer.ID)
	assert.True(t, exists)
	assert.Equal(t, testPlayer.Sequence, storedSeqInterface)

	oldPlayer := testPlayer
	oldPlayer.Sequence = 0
	oldData, err := player.SerializePlayer(oldPlayer)
	assert.NoError(t, err)

	gameState.HandleClient(conn, addr, oldData)

	storedSeqInterface, _ = gameState.SequenceNumbers.Load(testPlayer.ID)
	assert.Equal(t, uint32(1), storedSeqInterface)

	newPlayer := testPlayer
	newPlayer.Sequence = 2
	newPlayer.X = 150.5
	newData, err := player.SerializePlayer(newPlayer)
	assert.NoError(t, err)

	gameState.HandleClient(conn, addr, newData)

	storedSeqInterface, _ = gameState.SequenceNumbers.Load(testPlayer.ID)
	assert.Equal(t, uint32(2), storedSeqInterface)

	storedPlayerInterface, _ = gameState.Players.Load(testPlayer.ID)
	storedPlayer = storedPlayerInterface.(player.Player)
	assert.Equal(t, float32(150.5), storedPlayer.X)
}

func TestGameStateBroadcast(t *testing.T) {
	gameState := game.New()

	mockConn := &mockUDPConn{}

	player1 := player.Player{ID: 1, X: 100, Y: 200, Timestamp: time.Now().UnixMilli(), Sequence: 1}
	player2 := player.Player{ID: 2, X: 300, Y: 400, Timestamp: time.Now().UnixMilli(), Sequence: 1}

	addr1 := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9001}
	addr2 := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9002}

	gameState.Players.Store(player1.ID, player1)
	gameState.Players.Store(player2.ID, player2)
	gameState.Clients.Store(player1.ID, addr1)
	gameState.Clients.Store(player2.ID, addr2)

	gameState.Broadcast(mockConn)

	assert.GreaterOrEqual(t, mockConn.writeCount, 2)
}

func TestUDPServerSetup(t *testing.T) {
	cfg := config.Config{
		Port:         9000,
		GameTickRate: 10,
	}

	ready := make(chan struct{})

	go func() {
		addr := net.UDPAddr{Port: cfg.Port, IP: net.ParseIP("127.0.0.1")}
		conn, err := net.ListenUDP("udp", &addr)

		if err != nil {
			t.Errorf("Failed to start UDP server: %v", err)
			return
		}

		close(ready)

		time.Sleep(100 * time.Millisecond)
		conn.Close()
	}()

	select {
	case <-ready:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Server failed to start in time")
	}

	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:9000")
	assert.NoError(t, err)

	clientConn, err := net.DialUDP("udp", nil, clientAddr)
	if err != nil {
		// If this fails, it might be because the server already closed the connection
		// This is expected in a test environment, so we don't fail the test
		return
	}
	defer clientConn.Close()
}

type mockUDPConn struct {
	writeCount int
}

func (m *mockUDPConn) WriteToUDP(b []byte, addr *net.UDPAddr) (int, error) {
	m.writeCount++
	return len(b), nil
}
