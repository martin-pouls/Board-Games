# WebSocket review
- Half-dead sockets on phones (medium): `reconnectNow` in `shared/frontend/src/connection.svelte.ts` does nothing while `this.ws` is set, so after sleep or a network switch the socket looks open and the next action fails with `timeout` after 10s. Fix: on `visibilitychange` to visible after being hidden for a while, close and reopen the socket.
- No per-socket command rate limit (low): every command takes the game lock and loads from SQLite, rejected ones too. Bounded by the 4 KB read limit and the per-IP socket cap. Add a small token bucket per connection in `wskit`.
- Compression (low, unverified): `CompressionMode` is never set. Check the coder/websocket 1.8.15 default and consider enabling it, snapshots are repetitive JSON.
- Minor:
  - A hello with the same revision as a newer broadcast replaces `lastEvents` with `[]`, which can swallow an animation (narrow race).
  - A command applied just before the socket drops shows `disconnected`. A retry is rejected as illegal, which can show an error toast.
  - A second tab of the same player triggers a presence broadcast although the online list is unchanged.

# Minor
- Change favicon for start page, its still for card games
- Minor layout stuff
  - /board-games/ and the separate games (e.g. /the-game/) have different layouts. the latter is wider so the top bar (with all games and dark / light mode) is different. Unify this
  - also for the top bar on mobile the dark / light icon is placed differently for the start page and the separate games
  - in general i would like a more unified behavior here for the top bar at least. maybe make it shared or so? a shared header bar where certain components can be switched on or off (e.g. home or all games button)
  - for the current games the not so wide layout on desktop as well is fine. for future games we will also need a wider layout. you can keep that in mind. but its also fine to build that when we need it
  - Similar as above - on the password page the sun icon for the light mode is mangled. is this the same icon as the other screens? does this screen use the same style and lib (shadecn) as the others? if not it should. again use the same header bar
- Can we keep the Datenschutzerklärung more minimal? Templates i find on the internet tend to be shorter. also how long do we actually keep data? can we optimize something there? regarding caddy logs, etc.?