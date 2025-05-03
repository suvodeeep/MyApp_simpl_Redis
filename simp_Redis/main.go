package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"simp_Redis/server"
	"simp_Redis/store"
)

// Config holds application configs
type Config struct {
	TCPAddress string
	DataDir    string
	MaxMemory  string
	LogLevel   string
}

// loadConfigFromEnv loads configuration from environment variables
func loadConfigFromEnv() Config {
	config := Config{
		TCPAddress: "127.0.0.1:6379",
		DataDir:    "/data",
		MaxMemory:  "128MB",
		LogLevel:   "info",
	}

	if val := os.Getenv("TCP_PORT"); val != "" {
		config.TCPAddress = "0.0.0.0:" + val
	}

	if val := os.Getenv("DATA_DIR"); val != "" {
		config.DataDir = val
	}

	if val := os.Getenv("MAX_MEMORY"); val != "" {
		config.MaxMemory = val
	}

	if val := os.Getenv("LOG_LEVEL"); val != "" {
		config.LogLevel = val
	}

	return config
}

func main() {

	config := loadConfigFromEnv()

	// Parsing command-line flags (override env vars)
	address := flag.String("address", config.TCPAddress, "TCP address to listen on (host:port)")
	dataDir := flag.String("data-dir", config.DataDir, "Directory for data persistence")
	maxMemory := flag.String("max-memory", config.MaxMemory, "Maximum memory to use")
	logLevel := flag.String("log-level", config.LogLevel, "Log level (debug, info, warn, error)")
	flag.Parse()

	// Set up logging
	log.SetPrefix("[simp_Redis] ")
	switch *logLevel {
	case "debug":
		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Llongfile)
	case "info":
		log.SetFlags(log.Ldate | log.Ltime)
	case "warn", "error":
		log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	}

	log.Printf("Starting simp_Redis with TCP on %s", *address)
	log.Printf("Data directory: %s, Max memory: %s", *dataDir, *maxMemory)

	if err := os.MkdirAll(*dataDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	kvStore := store.NewStore()

	tcpServer := server.NewServer(kvStore)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Using a wait group to ensure both servers shut down gracefully
	var wg sync.WaitGroup
	wg.Add(1)

	// Starting the TCP server in a goroutine
	go func() {
		defer wg.Done()
		if err := tcpServer.Start(*address); err != nil {
			log.Fatalf("Failed to start TCP server: %v", err)
		}
	}()

	sig := <-signalChan
	log.Printf("Received signal %v, shutting down server...", sig)

	timeout := 50 * time.Second
	log.Printf("Allowing %v for graceful shutdown", timeout)

	shutdownChan := make(chan struct{})
	go func() {
		if err := tcpServer.Stop(); err != nil {
			log.Printf("Error stopping TCP server: %v", err)
		}
		close(shutdownChan)
	}()

	select {
	case <-shutdownChan:
		log.Printf("Servers stopped gracefully")
	case <-time.After(timeout):
		log.Printf("Timeout reached, forcing shutdown")
	}

	log.Printf("Server stopped")
}
