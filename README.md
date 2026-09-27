# Real-Time Multiplayer Chess Engine & Client

A full-stack, real-time multiplayer chess application built with **Go** and **Flutter**. The project features an authoritative Go backend handling move validation, clocks, and WebSocket orchestration, coupled with a responsive Flutter client featuring custom board visuals, offline move previews, audio feedback, and seamless reconnection handling.

---

## Table of Contents

- [Overview](#overview)
- [System Architecture](#system-architecture)
- [Features](#features)
- [Repository Structure](#repository-structure)
- [Technology Stack](#technology-stack)
- [API & Protocol Reference](#api--protocol-reference)
  - [HTTP API](#http-api)
  - [WebSocket Protocol](#websocket-protocol)
- [Local Development & Setup](#local-development--setup)
  - [Backend Setup](#backend-setup)
  - [Client Setup](#client-setup)
  - [Network Configuration](#network-configuration)
- [License](#license)

---

## Overview

This project provides a complete end-to-end real-time chess platform. The client renders the UI, handles local gesture inputs, pre-computes legal moves client-side using `dartchess` for immediate visual feedback, and streams player actions over WebSockets. The Go backend acts as the single source of truth—validating every move, managing per-player chess clocks, handling draw/resignation workflows, and broadcasting state updates back to client sessions.

---

## System Architecture

```
                                 +------------------------+
                                 |  Flutter Client        |
                                 |  (UI, Local Rules,     |
                                 |   Audio, State Cubit)  |
                                 +-----------+------------+
                                             |
                                 +-----------+------------+
                                 | REST API  | WebSockets |
                                 +-----+-----+-----+------+
                                       |           |
                                       v           v
+-----------------------------------------------------------------------------------+
| Go Backend                                                                        |
|                                                                                   |
|  +---------------------+   +---------------------+   +-------------------------+  |
|  | HTTP Handlers       |   | WS Connection Manager|  | Session & Room Registry |  |
|  +----------+----------+   +----------+----------+   +------------+------------+  |
|             |                         |                           |               |
|             +-------------------------+---------------------------+               |
|                                       |                                           |
|                                       v                                           |
|                          +-------------------------+                              |
|                          | Game Room Manager       |                              |
|                          | - Room Lifecycle        |                              |
|                          | - Reconnection Backoff  |                              |
|                          +------------+------------+                              |
|                                       |                                           |
|                                       v                                           |
|                          +-------------------------+                              |
|                          | Engine & Clock          |                              |
|                          | - Move Validation       |                              |
|                          | - Ticking Timers        |                              |
|                          +-------------------------+                              |
+-----------------------------------------------------------------------------------+
```

---

## Features

- **Public Matchmaking**: Quick game queueing that drops matching players directly into active rooms.
- **Private Rooms**: Custom games via shareable 6-character room invite codes.
- **Authoritative Server Engine**: Server-side move validation using `github.com/corentings/chess` to guarantee fair play and complete state control.
- **Instant Client Feedback**: Client-side move calculations (`dartchess`) for responsive selection highlights and promotion dialogs prior to network confirmation.
- **Synchronized Chess Clocks**: High-precision server clocks synced with local client countdowns and low-time acoustic warnings (under 10 seconds).
- **Resilient Reconnections**: Automatic exponential backoff client reconnects paired with backend session retention during transient network failures.
- **In-Game Actions**: Full support for draw offers (offer, accept, decline), resignations, check/checkmate highlights, and move/capture audio triggers.
- **Cross-Platform Client**: Built with Flutter for Desktop, Mobile (iOS/Android), and Web targets.

---

## Repository Structure

```
chess/
├── backend/                  # Go HTTP & WebSocket Server
│   ├── main.go               # Entry point and route initialization
│   ├── game/                 # Game engine wrapper, rooms, clock, and manager
│   ├── server/               # HTTP handlers, WS client upgrade & message loop
│   ├── go.mod                # Go module dependencies
│   └── README.md             # Backend specific documentation
│
└── chess/                    # Flutter Client Application
    ├── assets/               # Piece SVGs and WAV sound effects
    ├── lib/
    │   ├── main.dart         # Entry point, routing, and theme definitions
    │   ├── chess/            # Local rule engine wrappers (dartchess integration)
    │   ├── cubits/           # Game Cubit & state management
    │   ├── models/           # Data models (Game State, Messages, Sessions)
    │   ├── screens/          # Home, Waiting Room, and Active Game screens
    │   ├── services/         # API HTTP Client, WS Service, Sound, Storage
    │   ├── theme/            # Styling, color palette, dark mode definitions
    │   └── widgets/          # Board grid, pieces, action bars, modal overlays
    ├── pubspec.yaml          # Flutter dependencies
    └── README.md             # Client specific documentation
```

---

## Technology Stack

### Backend
- **Language**: Go 1.22+
- **HTTP / Router**: Standard `net/http` / `gorilla/mux`
- **WebSockets**: `gorilla/websocket`
- **Chess Engine**: `github.com/corentings/chess`

### Client
- **Framework**: Flutter (Dart ^3.11.4)
- **State Management**: `flutter_bloc` / `Cubit`
- **Networking**: `http`, `web_socket_channel`
- **Rules & Utilities**: `dartchess`, `shared_preferences`
- **Media & UI**: `audioplayers`, `flutter_svg`

---

## API & Protocol Reference

### HTTP API

| Method | Endpoint | Description | Request Body | Response Payload |
| :--- | :--- | :--- | :--- | :--- |
| `POST` | `/games/quick` | Join or create a public match | `{"nickname": "string"}` | `GameSession` (`game_id`, `session_id`, `player_id`) |
| `POST` | `/games/private` | Create a private room | `{"nickname": "string"}` | `GameSession` + `code` |
| `POST` | `/games/private/join` | Join a private room with code | `{"nickname": "string", "code": "string"}` | `GameSession` + `code` |
| `DELETE` | `/games` | Leave game / abandon session | `?sessionId={sessionId}` | `200 OK` |

### WebSocket Protocol

**Endpoint**: `ws://{host}:{port}/games/{gameId}/ws?sessionId={sessionId}`

#### Client $\rightarrow$ Server Events

```json
// Submit Move
{ "type": "move", "payload": { "from": "e2", "to": "e4", "promotion": "q" } }

// Offer Draw
{ "type": "offer_draw" }

// Respond to Draw
{ "type": "respond_draw", "payload": { "accepted": true } }

// Resign Game
{ "type": "resign" }
```

#### Server $\rightarrow$ Client Events

```json
// Authoritative Game State Broadcast
{
  "type": "game_state",
  "payload": {
    "game_id": "string",
    "white": { "id": "string", "nickname": "string" },
    "black": { "id": "string", "nickname": "string" },
    "fen": "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
    "white_time": 300000,
    "black_time": 300000,
    "active": 0,
    "status": "playing",
    "result": "none",
    "end_reason": "",
    "check": false,
    "draw_offered_by": null
  }
}
```

---

## Local Development & Setup

### Prerequisites

- **Go**: Version 1.22 or higher
- **Flutter SDK**: Dart SDK `^3.11.4`
- **Git**

### Backend Setup

1. Navigate to the backend folder:
   ```bash
   cd backend
   ```
2. Download dependencies:
   ```bash
   go mod download
   ```
3. Run the backend server (starts on `port 8080` by default):
   ```bash
   go run .
   ```

### Client Setup

1. Navigate to the client folder:
   ```bash
   cd chess
   ```
2. Install Flutter packages:
   ```bash
   flutter pub get
   ```
3. Run the application:
   ```bash
   flutter run
   ```

### Network Configuration

For local development across physical devices or emulators, update the server endpoint addresses in the Flutter client:

- **HTTP Base URL**: `lib/services/api/api_client.dart`
- **WebSocket Base URL**: `lib/services/game_socket_service.dart`

| Target Platform | Endpoint Address |
| :--- | :--- |

| **Physical Mobile Device** | `http://<YOUR_LOCAL_IP>:8080` / `ws://<YOUR_LOCAL_IP>:8080` |

---

## License

This repository is available under the [MIT License](LICENSE).