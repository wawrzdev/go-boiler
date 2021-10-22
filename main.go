package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/wawrzdev/go-boiler/configuration"
)

const (
	configFileName = "config"
	configFileType = "yaml"
)

func getDefaultConfiguration() *map[string]interface{} {
	return &map[string]interface{}{
		"BIND_ADDRESS":  ":9090",
		"API_NAME":      "API",
		"READ_TIMEOUT":  5,
		"WRITE_TIMEOUT": 10,
		"IDLE_TIMEOUT":  120,
	}
}

func getConfigurationPaths() (*[]string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("error getting current directory: %w", err)
	}
	return &[]string{"/etc/app", "$HOME/.app", ".", pwd}, nil
}

func main() {
	l := log.New(os.Stdout, "Config: ", log.LstdFlags)

	// set default config and load configuration from file
	defaultConfig := getDefaultConfiguration()
	configPaths, err := getConfigurationPaths()
	if err != nil {
		l.Printf("Error loading default config paths: %w\n", err)
		os.Exit(1)
	}
	configuration.SetDefaultConfiguration(defaultConfig)
	cf, err := configuration.LoadConfiguration(configFileName, configFileType, configPaths)
	if err != nil {
		l.Printf("Error reading configuration file: %w\n", err)
		os.Exit(1)
	}
	l.SetPrefix(fmt.Sprintf("%s: ", cf.API_NAME))

	// create the handlers
	l.Println("Creating handlers")

	// create a new serve mux and register handlers
	l.Println("Creating serve mux and registering handlers")
	sm := mux.NewRouter()

	// create a server
	s := http.Server{
		Addr:         cf.Server.BIND_ADDRRESS,               // configure the bind address
		Handler:      sm,                                    // set the default handler
		ErrorLog:     l,                                     // set the logger for the server
		ReadTimeout:  cf.Server.READ_TIMEOUT * time.Second,  // max time to read request from the client
		WriteTimeout: cf.Server.WRITE_TIMEOUT * time.Second, // max time to write response to the client
		IdleTimeout:  cf.Server.IDLE_TIMEOUT * time.Second,  // max time for connections using TCP Keep-Alive
	}

	sc, err := cf.Server.GetServerConfiguration()
	if err != nil {
		l.Printf("Configured server with default values\n")
	}
	l.Printf("Configured server with %v\n", sc)

	// start the server
	go func() {
		l.Printf("Starting server on %s\n", cf.Server.BIND_ADDRRESS)

		err := s.ListenAndServe()
		if err != nil {
			l.Printf("Error starting server: %s\n", err)
			os.Exit(1)
		}
	}()

	// trap signal and graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)

	// block until signal received
	sig := <-c
	l.Printf("Got signal: %s", sig)

	// gracefuly shutdown waiting max 30 seconds for operation completion
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	s.Shutdown(ctx)

}
