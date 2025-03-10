# Overview
A simple game client-server implementation to simulate 2D game environment where player can move the avatar within a border. The server implementation utilize UDP as the networking protocol.

# Goals
- Player can move their avatar and meet another player avatar in the game client.
- Server able to serve the client movement concurrently.
- Server able to handle disconnection of the player.
- Server able to broadcast players movements.
- Client can predict and interpolate the movement.

# Implementation
The server is implemented using UDP as the networking protocol.

UDP is used because we are aiming a low latency with lowest network cost. Compared with TCP, UDP has lower bandwith because UDP doesn't need to do a handshake on each of the connection. Besides that, since UDP isn't relying on the handshake, the packets send through UDP might loss in the transmit, but we managed to handle this in the client side by adding an interpolation and counting the sequence of received packet in the server.

The flow of the server:
1. Server opens UDP connection.
2. On each connection with the client, the server will spawn a new goroutine to handle the client connection separately.
3. Outdated packet will be ignored to not causing a bad experience to the client.
4. Server will monitor the disconnection of the clients for each 5 seconds.
5. Server will broadcast the clients position to another connected clients within the server.

The flow of the client:
1. Client connect to the server using UDP connection.
2. The client renders, and updates the board and also handling the packet send for the player.
3. The client predicts the position of the other clients by calculating the last position and time different from the last update.
4. The client handle incoming player or other client update separately using a goroutine. During the update, client reconcile the other player location based on the sequence and calculate the update time for the position interpolation if necessary.

# Future Improvement
Since this server is a simple game server, in the future, we can consider to add some feature for scalability.

## Server
1. A consensus algorithm to manage each instance of the server, but this can causing an additional latency, so this addition is depending on the game type.
2. Add authentication on the server.
3. Save the client state using a database.

## Client
1. Add dead reckoning to handle bad connection from the client/other clients.
