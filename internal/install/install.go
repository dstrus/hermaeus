// Package install finds the player's SkyrimNet folder: the one containing
// data, config, and logs, which may sit inside a Skyrim install, an MO2
// instance's overwrite folder, or anywhere the player chooses.
package install

import (
	"os"
	"path/filepath"
	"strings"

	"hermaeus/internal/apperr"
)

// Find looks for a SkyrimNet folder (one with data/*.db) in and below the
// places Hermaeus is likely to be started from: next to the program, or the
// current directory, which may be the SkyrimNet folder itself, a Skyrim
// install, or an MO2 instance. It returns "" if there isn't one.
func Find() string {
	var roots []string
	if exe, err := os.Executable(); err == nil {
		roots = append(roots, filepath.Dir(exe))
	}
	if wd, err := os.Getwd(); err == nil {
		roots = append(roots, wd)
	}
	return findIn(roots)
}

func findIn(roots []string) string {
	suffixes := []string{
		"",
		"SkyrimNet",
		filepath.Join("SKSE", "Plugins", "SkyrimNet"),
		filepath.Join("Data", "SKSE", "Plugins", "SkyrimNet"),
		filepath.Join("overwrite", "SKSE", "Plugins", "SkyrimNet"),
	}
	for _, root := range roots {
		for _, suffix := range suffixes {
			dir := filepath.Join(root, suffix)
			if len(FilesWithSuffix(filepath.Join(dir, "data"), ".db")) > 0 {
				return dir
			}
		}
	}
	return ""
}

// Check returns the absolute path of the SkyrimNet folder dir names, or an
// error saying why it isn't one. A path to the folder's data subfolder is
// accepted too, since that's where the databases are.
func Check(dir string) (string, error) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return "", apperr.New(apperr.Invalid, "Enter the path of your SkyrimNet folder")
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", apperr.New(apperr.Invalid, "Invalid folder: %v", err)
	}
	if !IsDir(abs) {
		return "", apperr.New(apperr.Invalid, "Folder not found: %s", abs)
	}
	if !IsDir(filepath.Join(abs, "data")) && strings.EqualFold(filepath.Base(abs), "data") {
		abs = filepath.Dir(abs)
	}
	if !IsDir(filepath.Join(abs, "data")) {
		return "", apperr.New(apperr.Invalid, "%s has no data folder. Choose the SkyrimNet folder (it contains data, config, and logs).", abs)
	}
	return abs, nil
}

// FilesWithSuffix lists the names of the files in dir ending in suffix
// (ignoring case). Unlike filepath.Glob it reads dir literally, so MO2 folder
// names such as "[NoDelete] SkyrimNet" aren't taken as patterns.
func FilesWithSuffix(dir, suffix string) []string {
	entries, _ := os.ReadDir(dir)
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), strings.ToLower(suffix)) {
			names = append(names, e.Name())
		}
	}
	return names
}

func IsDir(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && fi.IsDir()
}

func IsFile(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && fi.Mode().IsRegular()
}
