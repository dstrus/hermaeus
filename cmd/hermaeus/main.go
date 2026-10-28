// Hermaeus is an unofficial dashboard for SkyrimNet that works while Skyrim
// isn't running. It starts a small web server on this machine and opens it
// in the browser.
package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"hermaeus/internal/gamecheck"
	"hermaeus/internal/install"
	"hermaeus/internal/layers"
	"hermaeus/internal/prefs"
	"hermaeus/internal/server"
	"hermaeus/internal/store"
	"hermaeus/web"
)

// version is set when building a release, from the git tag.
var version = "dev"

func main() {
	log.SetFlags(log.Ltime)
	dir := flag.String("dir", "", "SkyrimNet folder, the one containing data, config, and logs (with MO2 usually overwrite\\SKSE\\Plugins\\SkyrimNet)")
	port := flag.Int("port", 8765, "port to listen on; if it's taken, the next nine are tried, then any free port")
	backups := flag.String("backups", "", "folder for backups (default: hermaeus-backups in the SkyrimNet folder)")
	modDir := flag.String("mod-dir", "", "SkyrimNet's mod folder, for its default bios when using Mod Organizer 2 (usually found automatically)")
	noBrowser := flag.Bool("no-browser", false, "don't open Hermaeus in a browser")
	showVersion := flag.Bool("version", false, "print Hermaeus's version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Println("Hermaeus", version)
		return
	}

	st := store.New(*backups, gamecheck.Running)
	remembered := prefs.Load(prefs.Path())
	saved := remembered.Get()
	switch {
	case *dir != "":
		if _, err := st.SetDir(*dir); err != nil {
			exit("%v", err)
		}
	case saved.SkyrimNetDir != "" && install.IsDir(saved.SkyrimNetDir):
		st.SetDir(saved.SkyrimNetDir)
	default:
		if found := install.Find(); found != "" {
			st.SetDir(found)
		}
	}

	library := layers.New(st.Dir)
	switch {
	case *modDir != "":
		if _, err := library.SetModDir(*modDir); err != nil {
			exit("%v", err)
		}
	case saved.ModDir != "" && install.IsDir(saved.ModDir):
		library.SetModDir(saved.ModDir)
	}

	ports := make([]int, 10)
	for i := range ports {
		ports[i] = *port + i
	}
	ln, running, err := server.Listen(ports)
	if running != "" {
		fmt.Printf("Hermaeus is already running at %s, so this opens it instead of starting a second copy.\n", running)
		if *dir != "" || *modDir != "" || *backups != "" {
			fmt.Println("To use different options, close that copy first, then start Hermaeus again.")
		}
		if !*noBrowser {
			openBrowser(running)
		}
		return
	}
	if err != nil {
		exit("Could not start Hermaeus: %v", err)
	}
	actualPort := ln.Addr().(*net.TCPAddr).Port
	handler, err := server.New(server.Options{
		Store:   st,
		Library: library,
		Prefs:   remembered,
		Token:   sessionToken(),
		Port:    actualPort,
		Files:   web.Files(),
		Version: version,
	})
	if err != nil {
		exit("%v", err)
	}
	srv := &http.Server{Handler: handler, ReadHeaderTimeout: 10 * time.Second}

	url := fmt.Sprintf("http://127.0.0.1:%d/", actualPort)
	fmt.Println()
	if actualPort != *port {
		fmt.Printf("Port %d is in use by another program, so Hermaeus is using port %d.\n", *port, actualPort)
	}
	fmt.Printf("Hermaeus %s (unofficial dashboard for SkyrimNet) running at %s\n", version, url)
	if d := st.Dir(); d == "" {
		fmt.Println("SkyrimNet folder: not found yet; choose it in the browser, or pass -dir")
	} else {
		fmt.Printf("SkyrimNet folder: %s\n", d)
		fmt.Printf("Backups go to:    %s\n", st.BackupDir())
	}
	fmt.Println("Close this window or press Ctrl+C to stop.")
	fmt.Println()
	if !*noBrowser {
		openBrowser(url)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		srv.Shutdown(shutdown)
	}()
	if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
		exit("%v", err)
	}
}

// sessionToken is a new random token for each run. During development the
// Vite dev server has to send the same one, so HERMAEUS_DEV_TOKEN sets it.
func sessionToken() string {
	if t := os.Getenv("HERMAEUS_DEV_TOKEN"); t != "" {
		return t
	}
	return rand.Text()
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		fmt.Printf("Open %s in your browser.\n", url)
	}
}

// exit reports a startup error. On Windows the console window closes as soon
// as the program ends when it was started by double-clicking, so wait for Enter.
func exit(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", a...)
	if runtime.GOOS == "windows" {
		fmt.Fprint(os.Stderr, "Press Enter to close.")
		bufio.NewReader(os.Stdin).ReadString('\n')
	}
	os.Exit(1)
}
