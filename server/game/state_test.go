package game

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/zainokta/client-server-multiplayer/server/player"
)

func TestNewGameState(t *testing.T) {
	gs := New()

	var count int
	gs.Players.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	assert.Equal(t, 0, count, "Players map should be empty")

	count = 0
	gs.Clients.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	assert.Equal(t, 0, count, "Clients map should be empty")

	count = 0
	gs.SequenceNumbers.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	assert.Equal(t, 0, count, "SequenceNumbers map should be empty")
}

func TestHandleClientWithInvalidData(t *testing.T) {
	gs := New()
	conn := &mockUDPConn{}
	addr := &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 9000,
	}

	invalidData := []byte(nil)
	gs.HandleClient(conn, addr, invalidData)

	var count int
	gs.Players.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	assert.Equal(t, 0, count, "No player should be added with invalid data")
}

func TestMonitorDisconnections(t *testing.T) {
	gs := New()
	conn := &mockUDPConn{}
	addr := &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 9000,
	}

	now := time.Now().UnixMilli()

	activePlayer := player.Player{
		ID:        1,
		X:         100,
		Y:         200,
		Timestamp: now,
		Sequence:  1,
	}

	inactivePlayer := player.Player{
		ID:        2,
		X:         300,
		Y:         400,
		Timestamp: now - (DisconnectTimer.Milliseconds() + 1000),
		Sequence:  1,
	}

	gs.Players.Store(activePlayer.ID, activePlayer)
	gs.Players.Store(inactivePlayer.ID, inactivePlayer)
	gs.Clients.Store(activePlayer.ID, &net.UDPAddr{})
	gs.Clients.Store(inactivePlayer.ID, &net.UDPAddr{})
	gs.SequenceNumbers.Store(activePlayer.ID, activePlayer.Sequence)
	gs.SequenceNumbers.Store(inactivePlayer.ID, inactivePlayer.Sequence)

	go func() {
		gs.MonitorDisconnections()
		time.Sleep(DisconnectTimer + 100*time.Millisecond)
	}()

	activePlayer.Timestamp = time.Now().UnixMilli()
	bytePlayer, err := player.SerializePlayer(activePlayer)
	assert.NoError(t, err)
	gs.HandleClient(conn, addr, bytePlayer)

	time.Sleep(DisconnectTimer + 200*time.Millisecond)

	_, activeExists := gs.Players.Load(activePlayer.ID)
	assert.True(t, activeExists, "Active player should still exist")

	_, inactiveExists := gs.Players.Load(inactivePlayer.ID)
	assert.False(t, inactiveExists, "Inactive player should be removed")
}

func TestBroadcastWithFailedWrite(t *testing.T) {
	gs := New()

	conn := &mockUDPConn{shouldFailWrite: true}

	testPlayer := player.Player{
		ID:        1,
		X:         100,
		Y:         200,
		Timestamp: time.Now().UnixMilli(),
		Sequence:  1,
	}

	addr := &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 9000,
	}

	gs.Players.Store(testPlayer.ID, testPlayer)
	gs.Clients.Store(testPlayer.ID, addr)

	gs.Broadcast(conn)

	_, exists := gs.Clients.Load(testPlayer.ID)
	assert.False(t, exists, "Client should be removed after failed write")
}

func TestHandleClientSequenceNumberHandling(t *testing.T) {
	gs := New()
	conn := &mockUDPConn{}
	addr := &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 9000,
	}

	initialPlayer := player.Player{
		ID:        1,
		X:         100,
		Y:         200,
		Timestamp: time.Now().UnixMilli(),
		Sequence:  5,
	}

	initialData, _ := player.SerializePlayer(initialPlayer)
	gs.HandleClient(conn, addr, initialData)

	storedSeq, _ := gs.SequenceNumbers.Load(initialPlayer.ID)
	assert.Equal(t, uint32(5), storedSeq.(uint32))

	oldPlayer := initialPlayer
	oldPlayer.Sequence = 3
	oldPlayer.X = 999

	oldData, _ := player.SerializePlayer(oldPlayer)
	gs.HandleClient(conn, addr, oldData)

	storedSeq, _ = gs.SequenceNumbers.Load(initialPlayer.ID)
	assert.Equal(t, uint32(5), storedSeq.(uint32))

	storedPlayerInterface, _ := gs.Players.Load(initialPlayer.ID)
	storedPlayer := storedPlayerInterface.(player.Player)
	assert.Equal(t, float32(100), storedPlayer.X)

	newPlayer := initialPlayer
	newPlayer.Sequence = 7
	newPlayer.X = 150

	newData, _ := player.SerializePlayer(newPlayer)
	gs.HandleClient(conn, addr, newData)

	storedSeq, _ = gs.SequenceNumbers.Load(initialPlayer.ID)
	assert.Equal(t, uint32(7), storedSeq.(uint32))

	storedPlayerInterface, _ = gs.Players.Load(initialPlayer.ID)
	storedPlayer = storedPlayerInterface.(player.Player)
	assert.Equal(t, float32(150), storedPlayer.X)
}

// Mock UDP connection for testing
type mockUDPConn struct {
	shouldFailWrite bool
}

func (m *mockUDPConn) WriteToUDP(b []byte, addr *net.UDPAddr) (int, error) {
	if m.shouldFailWrite {
		return 0, net.ErrClosed
	}
	return len(b), nil
}
