package main

import (
	"embed"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"fnos-terminal/pkg/gateway"
	"fnos-terminal/pkg/ws"
)

//go:embed all:dist
var embeddedDist embed.FS

func main() {
	socketPath := flag.String("socket", "", "Unix Domain Socket path for fnOS gateway (e.g. /var/apps/terminal/target/app.sock)")
	port := flag.Int("port", 5001, "TCP port for local development fallback")
	prefix := flag.String("prefix", "/app/terminal", "Gateway URL prefix")
	devMode := flag.Bool("dev", false, "Enable development mode (bypasses admin auth)")
	flag.Parse()

	cleanPrefix := strings.TrimSuffix(*prefix, "/")
	if cleanPrefix == "" {
		cleanPrefix = "/app/terminal"
	}

	distSub, err := fs.Sub(embeddedDist, "dist")
	if err != nil {
		log.Fatalf("Failed to load embedded dist fs: %v", err)
	}

	indexFile, err := distSub.Open("index.html")
	if err != nil {
		log.Fatalf("Failed to open embedded index.html: %v", err)
	}
	indexBytes, err := io.ReadAll(indexFile)
	if closeErr := indexFile.Close(); closeErr != nil {
		log.Printf("[Warning] Error closing index.html: %v", closeErr)
	}
	if err != nil {
		log.Fatalf("Failed to read embedded index.html: %v", err)
	}

	mux := http.NewServeMux()

	// 1. Health check endpoint
	mux.HandleFunc(cleanPrefix+"/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, writeErr := w.Write([]byte(`{"status":"ok"}`)); writeErr != nil {
			log.Printf("[HTTP] Failed writing health response: %v", writeErr)
		}
	})

	// 2. WebSocket Terminal endpoint
	termHandler := ws.NewTerminalHandler(*devMode)
	authMiddleware := gateway.Middleware(true, *devMode)
	mux.Handle(cleanPrefix+"/ws", authMiddleware(termHandler))

	// 3. Static Web Assets Handler (Zero-Redirect Guarantee)
	fileServer := http.FileServer(http.FS(distSub))

	serveIndex := func(w http.ResponseWriter) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.WriteHeader(http.StatusOK)
		if _, writeErr := w.Write(indexBytes); writeErr != nil {
			log.Printf("[HTTP] Failed writing index.html: %v", writeErr)
		}
	}

	staticHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqPath := strings.TrimPrefix(r.URL.Path, cleanPrefix)
		reqPath = strings.TrimPrefix(reqPath, "/")

		// Request for root, empty, or index.html -> direct serve index.html (NEVER redirect)
		if reqPath == "" || reqPath == "index.html" {
			serveIndex(w)
			return
		}

		// Check if file exists in embedded dist
		if f, openErr := distSub.Open(reqPath); openErr == nil {
			if closeErr := f.Close(); closeErr != nil {
				// handled
			}
			// Set explicit content types for assets
			if strings.HasSuffix(reqPath, ".js") {
				w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
			} else if strings.HasSuffix(reqPath, ".css") {
				w.Header().Set("Content-Type", "text/css; charset=utf-8")
			} else if strings.HasSuffix(reqPath, ".png") {
				w.Header().Set("Content-Type", "image/png")
			} else if strings.HasSuffix(reqPath, ".svg") {
				w.Header().Set("Content-Type", "image/svg+xml")
			}
			r.URL.Path = "/" + reqPath
			fileServer.ServeHTTP(w, r)
			return
		}

		// SPA fallback: return index.html
		serveIndex(w)
	})

	// Log all incoming requests for transparency
	loggingMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("[HTTP] %s %s (Admin=%s, User=%s)", r.Method, r.URL.Path, r.Header.Get("X-Trim-Isadmin"), r.Header.Get("X-Trim-Username"))
			next.ServeHTTP(w, r)
		})
	}

	// Mount both prefix and prefix/ to staticHandler to completely eliminate Go's internal 301 redirects
	mux.Handle(cleanPrefix, loggingMiddleware(staticHandler))
	mux.Handle(cleanPrefix+"/", loggingMiddleware(staticHandler))
	mux.Handle("/", loggingMiddleware(staticHandler))

	var listener net.Listener

	if *socketPath != "" {
		if rmErr := os.Remove(*socketPath); rmErr != nil && !os.IsNotExist(rmErr) {
			log.Printf("[Socket] Notice: removing existing socket file: %v", rmErr)
		}

		l, listenErr := net.Listen("unix", *socketPath)
		if listenErr != nil {
			log.Fatalf("[Socket] Failed to listen on unix socket %s: %v", *socketPath, listenErr)
		}
		listener = l
		log.Printf("[Server] fnOS Terminal listening on Unix Socket: %s (Prefix: %s)", *socketPath, cleanPrefix)

		if chmodErr := os.Chmod(*socketPath, 0666); chmodErr != nil {
			log.Printf("[Socket] Warning: failed to chmod socket file: %v", chmodErr)
		}
	} else {
		addr := fmt.Sprintf(":%d", *port)
		l, listenErr := net.Listen("tcp", addr)
		if listenErr != nil {
			log.Fatalf("[Server] Failed to listen on %s: %v", addr, listenErr)
		}
		listener = l
		log.Printf("[Server] fnOS Terminal listening at http://localhost%s%s/ (DevMode=%v)", addr, cleanPrefix, *devMode)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("[Server] Received signal %v, terminating...", sig)
		if closeErr := listener.Close(); closeErr != nil {
			log.Printf("[Server] Error closing listener: %v", closeErr)
		}
		if *socketPath != "" {
			if rmErr := os.Remove(*socketPath); rmErr != nil && !os.IsNotExist(rmErr) {
				log.Printf("[Socket] Error cleaning up socket on shutdown: %v", rmErr)
			}
		}
		os.Exit(0)
	}()

	server := &http.Server{Handler: mux}
	if serveErr := server.Serve(listener); serveErr != nil && serveErr != http.ErrServerClosed {
		log.Fatalf("[Server] Server stopped with error: %v", serveErr)
	}
}
