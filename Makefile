build-all:
	@echo "Building client and server..."
	@cp ./client/.env.example ./client/.env
	@cp ./server/.env.example ./server/.env
	@go build -o client/client ./client &
	@docker build -t multiplayer-server -f ./server/Dockerfile ./server/
	@wait

PORT=8000

run-server: 
	docker run -p $(PORT):$(PORT) -d -t multiplayer-server:latest

run: build-all run-server
	
stop:
	docker stop $(docker ps -q --filter ancestor=multiplayer-server)
