// Package gamecheck reports whether Skyrim is running. SkyrimNet keeps its
// database open while the game runs, so Hermaeus refuses to save then.
package gamecheck

import "strings"

// Executables that mean Skyrim is running.
var skyrimExes = []string{"skyrimse.exe", "skyrimvr.exe", "skyrim.exe"}

func isSkyrimExe(name string) bool {
	name = strings.ToLower(name)
	for _, exe := range skyrimExes {
		if name == exe {
			return true
		}
	}
	return false
}
