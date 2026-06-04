package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"
)

const protocolVersion = "0.2"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	mode := os.Args[1]
	os.Args = os.Args[1:] // shift args for flags

	switch mode {
	case "serve":
		port := flag.Int("port", 8080, "Port to listen on")
		pass := flag.String("pass", "", "Password for E2EE (enables built-in host)")
		bufferMB := flag.Int("buffer", 1, "Max buffer size in MB (when --pass is set)")
		flag.Parse()
		if *pass != "" {
			// Combined mode: serve + share in one process
			runServeAndShare(*port, *pass, *bufferMB)
		} else {
			runServer(*port)
		}
	case "share":
		server := flag.String("server", "ws://localhost:8080/ws", "Relay server URL")
		pass := flag.String("pass", "", "Password for E2EE")
		bufferMB := flag.Int("buffer", 1, "Max buffer size in MB (applies to terminal ring buffer and file manager)")
		flag.Parse()
		if *pass == "" {
			fmt.Println("Error: --pass is required for E2EE")
			return
		}
		runHost(*server, *pass, *bufferMB)
	case "join":
		server := flag.String("server", "ws://localhost:8080/ws", "Relay server URL")
		sessionID := flag.String("id", "", "Session ID to join")
		pass := flag.String("pass", "", "Password for E2EE")
		name := flag.String("name", "", "Display name of the viewer")
		flag.Parse()
		if *sessionID == "" || *pass == "" {
			fmt.Println("Error: --id and --pass are required")
			return
		}
		runViewer(*server, *sessionID, *pass, *name)
	default:
		printUsage()
	}
}

// runServeAndShare starts the relay server, waits for it to be ready,
// then connects a host session to it — all in one process.
func runServeAndShare(port int, password string, bufferMB int) {
	// Start server in background
	go runServer(port)

	// Wait for server to be ready (poll TCP)
	addr := fmt.Sprintf("localhost:%d", port)
	for i := 0; i < 50; i++ {
		conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if err == nil {
			conn.Close()
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	// Connect host to the local server
	wsURL := fmt.Sprintf("ws://localhost:%d/ws", port)
	runHost(wsURL, password, bufferMB)
}

func printUsage() {
	fmt.Println("rmte - Remote Terminal Relay")
	fmt.Println("Usage:")
	fmt.Println("  rmte serve --port=8080 [--pass=\"secret\" --buffer=1]")
	fmt.Println("  rmte share --server=\"ws://...\" --pass=\"secret\" [--buffer=1]")
	fmt.Println("  rmte join  --server=\"ws://...\" --id=\"...\" --pass=\"secret\" [--name=\"name\"]")
	fmt.Println("")
	fmt.Println("  serve --pass combines serve + share in one command.")
}
