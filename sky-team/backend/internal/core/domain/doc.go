// Package domain contains the pure game model and rule engine for Sky Team (docs/rules.md): the
// lobby lifecycle, the control panel and scenario tables, the flight across seven rounds, the
// dice placement and its automatic resolutions, the modules and the special abilities, and the
// per-player view.
//
// It has no dependencies on transport, persistence or any other package in
// this module — only the standard library. All rule enforcement lives here.
package domain
