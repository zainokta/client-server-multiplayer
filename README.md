# Simple Client-Server 2D Multiplayer
Golang Simple Client Server for 2D Game Simulation. 

# Getting Started
The server is a simple server to simulate 2D game (player can move their avatar freely).
- Go (1.22 or above)
- Makefile (optional)
- Docker (optional)

## Running the Project
This commands below to run the client-server using docker and makefile.
1. Build the client and server.
```shell
make build-all
```
2. Run the server instance.
```shell
make run
```        
3. Run the client instance (can be spawned multiple times).
```shell
./client/client
``` 

Alternatively, if you don't have docker and/or makefile, you can run the project using these commands:
1. Copy the client environment variables.
```shell
cp ./client/.env.example ./client/.env
```
2. Copy the server environment variables.
```shell
cp ./server/.env.example ./server/.env
```
3. Run the server instance.
```shell
go run ./server/main.go
```
4.  Run the client instance (can be spawned multiple times).
```shell
go run ./client/main.go
```

Client can be created multiple times to simulate a multiplayer environment.

# Usage
1. Start the server.
2. Start the client(s).
3. Play the client by using w/a/s/d key then press enter to move the player.

# Server Testing
```shell
go test ./server/...
```

# Technical Specification
[Docs](SPECIFICATION.md)