// Package ports declares the interfaces at the edge of the core.
//
//   - Driving (inbound) ports: use-case interfaces the adapters call,
//     e.g. a GameService (create/join/start game, play card, end turn).
//   - Driven (outbound) ports: interfaces the core needs implemented,
//     e.g. GameRepository (persistence) and EventPublisher (push updates
//     to connected clients).
//
// Ports depend only on the domain package.
package ports
