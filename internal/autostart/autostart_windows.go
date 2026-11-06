//go:build windows

package autostart

import (
	"errors"
	"os"

	"golang.org/x/sys/windows/registry"
)

// runKeyPath is the per-user registry key Windows consults at logon.
const runKeyPath = `Software\Microsoft\Windows\CurrentVersion\Run`

// IsEnabled returns whether launch at login is currently enabled.
func IsEnabled() (bool, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.QUERY_VALUE)
	if err != nil {
		return false, err
	}
	defer key.Close()
	if _, _, err := key.GetStringValue(appName); err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Enable registers the app to start at user login.
func Enable() error {
	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	key, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()
	return key.SetStringValue(appName, `"`+execPath+`"`)
}

// Disable removes the app from login startup. Removing a missing entry is a no-op.
func Disable() error {
	key, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()
	if err := key.DeleteValue(appName); err != nil && !errors.Is(err, registry.ErrNotExist) {
		return err
	}
	return nil
}
