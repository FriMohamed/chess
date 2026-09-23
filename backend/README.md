# Chess Backend

A real-time multiplayer chess backend written in Go. This service handles game matchmaking, move validation, clock management, and real-time communication via WebSockets.

## Features

- **Matchmaking**: Support for Quick Games (public matchmaking) and Private Games (via invite codes).
- **Real-time Gameplay**: WebSocket-based communication for moves, game state updates, and clock synchronization.
- **Game Management**: Automated room cleanup, connection timeouts, and reconnection support.
- **Chess Engine**: Robust move validation and game rules powered by the `github.com/corentings/chess` engine.
- **Session Management**: Secure session handling to allow players to reconnect to their games.

## Architecture

The backend is divided into two main packages:

### 1. `game` Package
Responsible for the core chess logic and game room management.
- **Manager**: Orchestrates all active game rooms.
- **Room**: Handles synchronization between players, manages connection timers, and acts as a thread-safe wrapper around the game state.
- **Game**: Manages the chess board, move execution, and game outcomes.
- **Clock**: Tracks time for both players with automatic timeout detection.

### 2. `server` Package
Handles the networking layer.
- **HTTP Handlers**: For creating games, joining private rooms, and session management.
- **WebSocket Handlers**: Manages the real-time message loop for active games.
- **Client**: Wraps the WebSocket connection and handles concurrent writes.

## API Endpoints

### HTTP API
- `POST /games/quick`: Join or create a public matchmaking room.
- `POST /games/private`: Create a new private game with an invite code.
- `POST /games/private/join`: Join an existing private game using a code.
- `DELETE /games`: Leave a game and end the session.
<!-- - `GET /games`: (Diagnostic) List all open games. -->

### WebSocket API
- `GET /games/{gameID}/ws?sessionId={sessionId}`: Establish a real-time connection for an active game.

## Getting Started

### Prerequisites
- Go 1.22+ (Note: `go.mod` specifies 1.25, but 1.22+ should be compatible).

### Installation
1. Clone the repository.
2. Install dependencies:
   ```bash
   go mod download
   ```
3. Run the server:
   ```bash
   go run .
   ```