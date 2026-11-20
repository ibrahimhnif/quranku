//go:build e2e

package main

// Builds with the e2e tag let another process drive the app, as an agent or
// an end-to-end test does: MYGO_AUTOMATION names its socket.
import _ "github.com/egoist/mygo/plugins/automation"
