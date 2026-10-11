# Technical
- Access layer: shared password, plan in [ACCESS-LAYER.md](ACCESS-LAYER.md).

## WebSocket review
- Half-dead sockets on phones (medium): `reconnectNow` in `shared/frontend/src/connection.svelte.ts` does nothing while `this.ws` is set, so after sleep or a network switch the socket looks open and the next action fails with `timeout` after 10s. Fix: on `visibilitychange` to visible after being hidden for a while, close and reopen the socket.
- No per-socket command rate limit (low): every command takes the game lock and loads from SQLite, rejected ones too. Bounded by the 4 KB read limit and the per-IP socket cap. Add a small token bucket per connection in `wskit`.
- Compression (low, unverified): `CompressionMode` is never set. Check the coder/websocket 1.8.15 default and consider enabling it, snapshots are repetitive JSON.
- Minor:
  - A hello with the same revision as a newer broadcast replaces `lastEvents` with `[]`, which can swallow an animation (narrow race).
  - A command applied just before the socket drops shows `disconnected`. A retry is rejected as illegal, which can show an error toast.
  - A second tab of the same player triggers a presence broadcast although the online list is unchanged.
