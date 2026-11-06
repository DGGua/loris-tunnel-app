// Package autostart manages launch-at-login for the current user.
// Windows uses the per-user registry Run key; macOS uses a launchd agent.
package autostart

const (
	appName        = "loris-tunnel"
	appDisplayName = "Loris Tunnel"
)
