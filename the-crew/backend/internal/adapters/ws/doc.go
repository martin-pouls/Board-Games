// Package ws is the WebSocket adapter for live gameplay. It implements the
// EventPublisher driven port (broadcasting per-player game views to
// connected clients) and accepts in-game commands as a driving adapter.
//
// The transport (connections, rooms, presence, read loop, close codes) is
// shared: shared/backend wskit. This package builds The Crew's view messages
// and dispatches The Crew's commands. Construct with NewHub and NewHandler;
// the Hub is passed to the service as its EventPublisher.
package ws
