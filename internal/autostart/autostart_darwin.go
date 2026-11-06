//go:build darwin

package autostart

import (
	"os"

	"github.com/emersion/go-autostart"
)

// newApp describes the launchd agent for the current user.
func newApp() *autostart.App {
	execPath, _ := os.Executable()
	if execPath == "" {
		execPath = "loris-tunnel"
	}
	return &autostart.App{
		Name:        appName,
		DisplayName: appDisplayName,
		Exec:        []string{execPath},
	}
}

// IsEnabled returns whether launch at login is currently enabled.
func IsEnabled() (bool, error) {
	return newApp().IsEnabled(), nil
}

// Enable registers the app to start at user login.
func Enable() error {
	return newApp().Enable()
}

// Disable removes the app from login startup.
func Disable() error {
	return newApp().Disable()
}
