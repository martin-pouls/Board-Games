// Package backend is the plumbing every game backend shares: the wire envelope, HTTP helpers
// and abuse limits, WebSocket transport, SQLite snapshot store, ID and token generation, a
// keyed mutex and the server bootstrap. It knows nothing about any game: no rules, no domain
// types. Games wire it up in their adapters and cmd/server; their core/domain never imports it.
//
//	wire       JSON envelope shared by REST and WS (error body, view/presence/ack/error messages)
//	httpkit    JSON helpers, bearer token, SPA handler, error mapping, router skeleton, rate limits
//	wskit      WebSocket connections, per-game hub and the endpoint (auth, presence, read loop)
//	sqlitedb   SQLite open + migrations, and the games/player_tokens snapshot store
//	idgen      game codes, player IDs, tokens and their hashes, seeded RNGs
//	lockmap    a per-key mutex
//	serverkit  env config, graceful HTTP serve, janitor
package backend
