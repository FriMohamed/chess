# Chess (Flutter Client)

A real-time multiplayer chess app built with Flutter. It pairs with the [Chess Backend](../backend/README.md) — a Go server that handles matchmaking, authoritative move validation, clocks, and live game state — and communicates with it over REST for matchmaking and WebSockets for gameplay.

The client renders the board, manages local interaction (selection, legal-move hints, promotion, clock countdown, sounds), and treats the backend as the single source of truth for game state.

---

## Table of Contents

- [Features](#features)
- [Screens & User Flow](#screens--user-flow)
- [Repository Layout](#repository-layout)
- [Project Structure](#project-structure)
- [Architecture](#architecture)
- [State Management](#state-management)
- [Networking](#networking)
  - [HTTP API](#http-api)
  - [WebSocket Protocol](#websocket-protocol)
  - [Connection Handling](#connection-handling)
- [Configuration (Local Development)](#configuration-local-development)
- [Game Logic](#game-logic)
- [UI, Theme & Sound](#ui-theme--sound)
- [Getting Started](#getting-started)
- [Testing](#testing)
- [Tech Stack](#tech-stack)
- [Related](#related)

---

## Features

- **Quick Game** — Public matchmaking that drops you into a game with a random opponent.
- **Private Games** — Create a game and share the generated invite code, or join an existing game by entering a code.
- **Real-time Gameplay** — Moves, clocks, draws, resignations, and game-over events sync instantly over a WebSocket.
- **Resilient Connection** — Automatic reconnection with exponential backoff, plus a "connection lost" overlay with a manual **Try Again**.
- **Chess Clock** — Per-player countdown ticking locally each second from server-provided times, with a low-time warning sound at 10 seconds.
- **Full Move Handling** — Piece selection, legal-move dots, check highlighting, and a pawn promotion picker.
- **Game Actions** — Offer / accept / decline draw offers, resign, and leave a game.
- **Sound Effects** — Move, capture, draw offer, and low-time sounds.
- **Persistent Nickname** — Your name is stored locally and editable from the home screen; a random chess-themed name is generated on first launch.
- **Dark UI** — A gold-on-charcoal theme built around a custom `ColorScheme`.

## Screens & User Flow

The app has three screens under `lib/screens/`.

### 1. Home (`home_screen.dart`)

- Displays the app title and an editable **player name** (`PlayerName` widget).
- **Quick Game** → navigates to the waiting screen in `GameMode.public`.
- **Private** → opens a bottom sheet with:
  - **Create Game** → waiting screen in `GameMode.privateCreate`.
  - **Join Game** → a dialog to enter a code, then waiting screen in `GameMode.privateJoin`.

> No network call happens on the home screen. It only chooses a mode; the waiting screen performs the actual API request.

### 2. Waiting (`game_waiting_screen.dart`)

- Reads the player's nickname, then calls the API matching the `GameMode`.
- For **privateCreate**, displays the backend-generated code with tap-to-copy.
- Reacts to the socket's `game_started` message by showing **OPPONENT FOUND** and then pushing the game screen.
- Shows an `ErrorOverlay` on failure (e.g. "invalid or full" code) and a confirm modal when trying to leave/cancel matchmaking.

### 3. Game (`game_screen.dart`)

- Owns the `GameSocketService` and creates a `GameCubit`.
- Renders two `PlayerBar`s (opponent on top, you on the bottom), the `ChessBoard`, and overlays:
  - `PromotionPicker` when a pawn reaches the last rank.
  - `GameFinishedOverlay` when the game ends (win / lose / draw with the reason).
  - Reconnecting / disconnected overlays driven by socket status.
- Intercepts back navigation with a leave-game confirmation (or goes home directly if the game is finished).

## Repository Layout

This project lives in a monorepo alongside its backend:

```
chess/
├── chess/       # Flutter client (this document)
└── backend/     # Go HTTP + WebSocket server (see backend/README.md)
```

| Path | Description |
| --- | --- |
| [`chess/`](.) | Flutter client application. |
| [`backend/`](../backend/README.md) | Go server: matchmaking, chess engine, clocks, WebSockets. |

## Project Structure

```
lib/
├── main.dart                      # App entry point, theme, initial route
├── chess/
│   └── chess_rules.dart           # Legal moves, promotion, check detection (dartchess)
├── cubits/
│   ├── game_cubit.dart            # Game state machine + socket message handling
│   └── game_screen_state.dart     # Immutable screen state + DrawOfferState enum
├── models/
│   ├── game_state.dart            # Authoritative game snapshot + enums
│   ├── game_session.dart          # gameId / sessionId / playerId
│   ├── game_message.dart          # Legacy message wrapper
│   ├── player.dart                # Player id + nickname
│   ├── quick_game_response.dart   # Quick-game API response
│   └── private_game_response.dart # Private-game API response (+ code)
├── screens/
│   ├── home_screen.dart           # Mode selection & nickname
│   ├── game_waiting_screen.dart   # Matchmaking / private code / waiting
│   └── game_screen.dart           # Board, player bars, overlays
├── services/
│   ├── api/
│   │   ├── api_client.dart        # Thin HTTP wrapper (POST / DELETE)
│   │   └── game_api_service.dart  # Game endpoints
│   ├── game_socket_service.dart   # WebSocket client + reconnection
│   ├── game_sound_service.dart    # audioplayers wrapper
│   └── player_name_service.dart   # shared_preferences nickname persistence
├── theme/
│   └── app_theme.dart             # Dark theme & color palette
└── widgets/
    ├── custom_button.dart         # Full-width styled button
    ├── confirm_modal.dart         # Reusable confirm/cancel dialog
    ├── error_overlay.dart         # Full-screen error with two actions
    ├── loading_dots.dart          # Animated "..." indicator
    ├── player_name.dart           # Inline-editable nickname
    └── game_screen/
        ├── chess_board.dart       # 8x8 grid built from FEN
        ├── chess_square.dart      # Square: selection, legal moves, check
        ├── chess_piece.dart       # SVG piece renderer
        ├── player_bar.dart        # Name, clock, draw/resign actions
        ├── promotion_picker.dart  # Q / R / B / N chooser
        └── game_finished_overlay.dart # Result + end reason
```

Assets live in `assets/`:

- `assets/pieces/white/` — `K Q R B N P` (uppercase, white pieces)
- `assets/pieces/black/` — `k q r b n p` (lowercase, black pieces)
- `assets/sounds/` — `move.wav`, `capture.wav`, `lowtime.wav`, `drawoffer.wav`, `turnswitch.wav`

## Architecture

The app follows a layered client architecture:

1. **Screens** handle widget composition and navigation.
2. **`GameCubit`** (business logic) derives UI state from `GameState` snapshots and interprets socket messages.
3. **Services** isolate side effects — HTTP, WebSockets, audio, and local storage.
4. **Models** are plain Dart classes that mirror the backend JSON payloads.
5. **`ChessRules`** uses the `dartchess` package for rule queries (legal moves, check, promotion detection) so the UI can give immediate feedback without a server round-trip.

The backend remains authoritative: the client optimistically renders selections/legal moves, but every move is sent to the server, and the resulting `game_state` broadcast replaces local state.

## State Management

State management uses [`flutter_bloc`](https://pub.dev/packages/flutter_bloc) with a `Cubit`.

### `GameCubit`

Created per game screen with the current `GameState`, the local `playerId`, and a shared `GameSocketService`.

- Subscribes to socket messages and applies authoritative updates via `_updateGame`.
- Runs a 1-second `Timer.periodic` clock that decrements the active player's remaining time locally and plays a low-time sound once per player at ≤ 10s. The timer only ticks while status is `playing`.
- Detects captures by comparing piece counts in the FEN before/after a move, to choose the capture vs. move sound.
- Derives `DrawOfferState` reactively from `game.drawOfferedBy` rather than tracking offers with local timers:
  - `none` — no offer.
  - `sent` — you offered.
  - `received` — the opponent offered (also plays the draw-offer sound).
  - `declined` — reserved.
- Exposes intents: `selectSquare`, `promote`, `offerDraw`, `respondToDraw`, `resign`. Commands are sent to the socket, and the UI waits for the server's state broadcast to confirm them.

### `GameScreenState`

Immutable snapshot containing the `GameState`, local `whiteTime`/`blackTime`, the selected square, the list of legal moves for the selection, pending promotion squares, and the `DrawOfferState`.

### Local state ownership

- `_GameScreenState` owns socket-status subscriptions and navigation guards.
- `GameWaitingScreen` owns the matchmaking lifecycle and the socket used until the game starts.

## Networking

### HTTP API

All REST calls go through `ApiClient` (`lib/services/api/api_client.dart`), which performs JSON `POST`/`DELETE` requests and throws on non-2xx responses.

| Method | Endpoint | Body | Response |
| --- | --- | --- | --- |
| `POST` | `/games/quick` | `{ "nickname": string }` | `GameSession` (`game_id`, `session_id`, `player_id`) |
| `POST` | `/games/private` | `{ "nickname": string }` | `GameSession` + `code` |
| `POST` | `/games/private/join` | `{ "nickname": string, "code": string }` | `GameSession` + `code` |
| `DELETE` | `/games?sessionId={sessionId}` | — | Empty (leave / end session) |

The endpoints are wrapped by `GameApiService` (`lib/services/api/game_api_service.dart`).

### WebSocket Protocol

A single WebSocket connection carries all real-time game traffic:

```
ws://{host}/games/{gameId}/ws?sessionId={sessionId}
```

Connection status is tracked with `GameSocketStatus`: `disconnected`, `connecting`, `connected`, `reconnecting`.

#### Client → Server

| Type | Payload | Meaning |
| --- | --- | --- |
| `move` | `{ "from": "e2", "to": "e4", "promotion"?: "q" }` | Make a move (`promotion` only for pawn promotion). |
| `resign` | — | Resign the game. |
| `offer_draw` | — | Offer a draw. |
| `respond_draw` | `{ "accepted": bool }` | Accept or decline an incoming draw offer. |

#### Server → Client

| Type | Payload | Interpreted as |
| --- | --- | --- |
| `game_started` | `GameState` | `GameStartedMessage` — the game begins. |
| `game_state` | `GameState` | `GameStateMessage` — authoritative snapshot. |
| `draw_offered` | `{ "player_id": string }` | `DrawOfferedMessage` — an opponent offered a draw. |
| `draw_declined` | — | `DrawDeclinedMessage` — a draw offer was declined. |
| `error` | `{ "code": string, "message": string }` | `ErrorMessage` — server-side error. |
| *(unknown)* | — | `UnknownMessage`. |

Messages are decoded into a sealed `GameSocketMessage` hierarchy and exposed via a broadcast stream.

#### `GameState` payload

| Field | Type | Notes |
| --- | --- | --- |
| `game_id` | string | Game identifier. |
| `white` / `black` | `Player` \| null | `{ "id": string, "nickname": string }`; null while waiting. |
| `fen` | string | Board position in FEN. |
| `white_time` / `black_time` | int | Remaining time in milliseconds. |
| `active` | int | `0` = white to move, `1` = black to move. |
| `status` | string | `waiting` \| `playing` \| `finished`. |
| `result` | string | `none` \| `white_wins` \| `black_wins` \| `draw`. |
| `end_reason` | string | e.g. `checkmate`, `stalemate`, `resignation`, `timeout`, `draw_agreement`, `insufficient_material`, `disconnect`, `quit`, `threefold_repetition`, `fifty_move_rule`, … |
| `check` | bool | Whether the side to move is in check. |
| `draw_offered_by` | string \| null | Player id that currently has a draw offer on the table. |

### Connection Handling

`GameSocketService` (`lib/services/game_socket_service.dart`) manages the socket lifecycle:

- Uses `web_socket_channel` and awaits `channel.ready` before marking itself `connected`.
- On unexpected close or error, schedules a reconnect using backoff delays of **2s, 3s, 5s, 8s, 10s**; after exhausting them it transitions to `disconnected`.
- Ignores stale events from superseded channels and respects intentional closes.
- `retry()` manually resets backoff and reconnects; `close()` / `dispose()` tear the connection and streams down.
- Exposes `messages` (game events), `errors` (transport errors), and `statusStream` (connection status) streams.

## Configuration (Local Development)

The backend URLs are currently **hardcoded** for local development testing:

| Setting | File | Default |
| --- | --- | --- |
| HTTP base URL | `lib/services/api/api_client.dart` | `http://192.168.11.108:8080` |
| WebSocket base URL | `lib/services/game_socket_service.dart` | `ws://192.168.11.108:8080` |

`192.168.11.108` is a LAN address used to test on a **physical device** against a backend running on the developer's machine on port `8080`. To run the app, point these at wherever your backend is reachable:

| Target | Use |
| --- | --- |
| Android emulator | `10.0.2.2:8080` (emulator alias for the host machine) |
| iOS simulator / desktop / web | `localhost:8080` or `127.0.0.1:8080` |
| Physical device | Your machine's LAN IP (e.g. `192.168.x.x:8080`), and ensure both are on the same network |

> When you move off local development, replace these constants with environment/config-driven values.

## Game Logic

`ChessRules` (`lib/chess/chess_rules.dart`) wraps the [`dartchess`](https://pub.dev/packages/dartchess) package:

- `legalMoves(fen, from)` — legal destination squares for a piece.
- `isPromotionMove(fen, from)` — true when a pawn sits on the promotion rank.
- `isCheck(fen)` / `checkedKingSquare(fen)` — check detection and the king's square for highlighting.

The board (`chess_board.dart`) parses the FEN position into an 8×8 array and mirrors the grid when playing as black so your pieces are always at the bottom. `chess_square.dart` styles squares by selection, legal-move availability, and check, and renders SVG pieces from `assets/pieces/`.

## UI, Theme & Sound

- **Theme** (`lib/theme/app_theme.dart`) — dark, gold-on-charcoal palette: background `#111315`, surface `#191C1F`, primary gold `#C9A96E`, error `#D96B6B`. Rounded inputs and buttons.
- **Board colors** — light squares `#F0D9B5`, dark squares `#B58863`; selection yellow, check red, legal moves green dots.
- **Animations** — selected pieces scale, `AnimatedContainer` highlights the active player's bar, and `LoadingDots` animates during waiting/reconnecting.
- **Sound** (`lib/services/game_sound_service.dart`) — `playMove`, `playCapture`, `playLowTime`, `playDrawOffer`, and `playTurnSwitch` mapped to the WAV assets in `assets/sounds/`.

## Getting Started

### Prerequisites

- [Flutter SDK](https://docs.flutter.dev/get-started/install) — the project targets Dart SDK `^3.11.4`.
- A running [Chess Backend](../backend/README.md) instance on port `8080`.

### Installation

1. Clone the repository.
2. From the `chess/` directory, install dependencies:
   ```bash
   flutter pub get
   ```
3. Configure the backend URL for your environment (see [Configuration](#configuration-local-development)).
4. Run the app:
   ```bash
   flutter run
   ```

### Platform support

Standard Flutter platform folders are included: `android/`, `ios/`, `linux/`, `macos/`, `web/`, and `windows/`.

## Testing

The project includes the default Flutter widget test scaffold at `test/widget_test.dart`. It currently contains the counter smoke test from the Flutter template and does not reflect the app's widgets yet. Run tests with:

```bash
flutter test
```

Linting uses `flutter_lints` (see `analysis_options.yaml`):

```bash
flutter analyze
```

## Tech Stack

| Package | Purpose |
| --- | --- |
| [`flutter_bloc`](https://pub.dev/packages/flutter_bloc) | State management (`GameCubit`). |
| [`http`](https://pub.dev/packages/http) | REST API calls for matchmaking / leaving games. |
| [`web_socket_channel`](https://pub.dev/packages/web_socket_channel) | Real-time WebSocket gameplay. |
| [`dartchess`](https://pub.dev/packages/dartchess) | Chess rules: legal moves, check, promotion. |
| [`shared_preferences`](https://pub.dev/packages/shared_preferences) | Persisting the player nickname. |
| [`audioplayers`](https://pub.dev/packages/audioplayers) | Game sound effects. |
| [`flutter_svg`](https://pub.dev/packages/flutter_svg) | Rendering SVG chess pieces. |
| [`cupertino_icons`](https://pub.dev/packages/cupertino_icons) | iOS-style icon font. |

## Related

- [Chess Backend](../backend/README.md) — Go server providing matchmaking, game rooms, the chess engine, clocks, and WebSockets.
