// Package ws is the WebSocket adapter for live gameplay. It implements the
// EventPublisher driven port (broadcasting per-player game views to
// connected clients) and accepts in-game commands as a driving adapter.
//
// Library: github.com/coder/websocket. Construct with NewHub and NewHandler;
// the Hub is passed to the service as its EventPublisher.
package ws
