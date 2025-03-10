package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/zainokta/client-server-multiplayer/client/config"
	"github.com/zainokta/client-server-multiplayer/client/player"

	_ "github.com/joho/godotenv/autoload"
)

var sequenceNumber uint32 = 0

const (
	width  = 20
	height = 10
)

func main() {
	cfg, err := env.ParseAs[config.Config]()
	if err != nil {
		fmt.Printf("%+v\n", err)
	}

	addr := net.UDPAddr{Port: cfg.Port, IP: net.ParseIP("127.0.0.1")}
	conn, err := net.DialUDP("udp", nil, &addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	go player.GetPlayerUpdate(conn)

	gameBoard := make([][]rune, height)
	for i := range gameBoard {
		gameBoard[i] = make([]rune, width)
		for j := range gameBoard[i] {
			gameBoard[i][j] = ' '
		}
	}

	inputChan := make(chan rune, 10)
	stopChan := make(chan struct{})

	reader := bufio.NewReader(os.Stdin)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stopChan:
				return
			default:
				input, err := reader.ReadString('\n')
				if err != nil {
					time.Sleep(100 * time.Millisecond)
					continue
				}

				input = strings.TrimSpace(strings.ToLower(input))
				if len(input) > 0 {
					inputChan <- rune(input[0])
				}
			}
		}
	}()

	fmt.Println("A simple 2D real-time environment")
	fmt.Println("Controls: W = Up, A = Left, S = Down, D = Right, Q = Quit")
	fmt.Println("Press Enter after each command")
	fmt.Println("Game starting...")
	time.Sleep(2 * time.Second)

	gameTicker := time.NewTicker(time.Second / time.Duration(cfg.GameTickRate))
	defer gameTicker.Stop()

	networkTicker := time.NewTicker(time.Second / time.Duration(cfg.GameTickRate/2))
	defer networkTicker.Stop()

	gamePlayer := player.New(width/2, height/2)

	lastUpdateTime := time.Now()
	playing := true

	var playerMutex sync.Mutex

	wg.Add(1)
	go func() {
		defer wg.Done()
		for playing {
			select {
			case <-stopChan:
				return

			case cmd := <-inputChan:
				playerMutex.Lock()
				moved := false
				switch cmd {
				case 'w':
					if gamePlayer.Y > 1 {
						gamePlayer.Y--
						moved = true
					}
				case 'a':
					if gamePlayer.X > 1 {
						gamePlayer.X--
						moved = true
					}
				case 's':
					if gamePlayer.Y < height-2 {
						gamePlayer.Y++
						moved = true
					}
				case 'd':
					if gamePlayer.X < width-2 {
						gamePlayer.X++
						moved = true
					}
				case 'q':
					playing = false
					close(stopChan)
				}
				playerMutex.Unlock()

				if moved {
					playerMutex.Lock()
					sequenceNumber++
					gamePlayer.Timestamp = time.Now().UnixMilli()
					gamePlayer.Sequence = sequenceNumber
					player.SendPlayerUpdate(conn, gamePlayer)
					lastUpdateTime = time.Now()
					playerMutex.Unlock()
				}

			case <-gameTicker.C:
				playerMutex.Lock()
				updateBoard(gameBoard, gamePlayer)
				clearScreen()
				renderGame(gameBoard)
				fmt.Printf("\nPlayer position: (%.2f, %.2f)\n", gamePlayer.X, gamePlayer.Y)
				fmt.Println("Enter move (w/a/s/d) or q to quit:")
				playerMutex.Unlock()

			case <-networkTicker.C:
				if time.Since(lastUpdateTime) > time.Second/time.Duration(cfg.GameTickRate) {
					playerMutex.Lock()
					sequenceNumber++
					gamePlayer.Timestamp = time.Now().UnixMilli()
					gamePlayer.Sequence = sequenceNumber
					player.SendPlayerUpdate(conn, gamePlayer)
					lastUpdateTime = time.Now()
					playerMutex.Unlock()
				}
			}
		}
	}()

	wg.Wait()
}

func updateBoard(board [][]rune, gamePlayer player.Player) {
	for i := range board {
		for j := range board[i] {
			board[i][j] = ' '
		}
	}

	for i := range height {
		for j := range width {
			if i == 0 || i == height-1 || j == 0 || j == width-1 {
				board[i][j] = '#'
			}
		}
	}

	px, py := int(gamePlayer.X), int(gamePlayer.Y)
	if px >= 0 && px < width && py >= 0 && py < height {
		board[py][px] = 'o'
	}

	player.Players.Range(func(_, value any) bool {
		otherPlayer, ok := value.(player.Player)
		if !ok {
			return true
		}

		predicted := player.PredictPosition(otherPlayer)
		if otherPlayer.ID != gamePlayer.ID {
			ox, oy := int(predicted.X), int(predicted.Y)
			if ox >= 0 && ox < width && oy >= 0 && oy < height {
				board[oy][ox] = 'X'
			}
		}

		player.Players.Store(otherPlayer.ID, predicted)

		return true
	})
}

func renderGame(board [][]rune) {
	for _, row := range board {
		for _, cell := range row {
			fmt.Printf("%c ", cell)
		}
		fmt.Println()
	}
}

func clearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}
