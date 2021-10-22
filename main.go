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
	banner = `
   _____          ____          _  _             
  / ____|        |  _ \        (_)| |            
 | |  __   ___   | |_) |  ___   _ | |  ___  _ __ 
 | | |_ | / _ \  |  _ <  / _ \ | || | / _ \| '__|
 | |__| || (_) | | |_) || (_) || || ||  __/| |   
  \_____| \___/  |____/  \___/ |_||_| \___||_|   
                                                 
                                                 
`
)

const (
	configFileName = "config"
	configFileType = "yaml"
)

func getDefaultConfiguration() *map[string]interface{} {
	return &map[string]interface{}{
		"API_NAME":             "API",
		"server.BIND_ADDRESS":  ":9090",
		"server.READ_TIMEOUT":  5,
		"server.WRITE_TIMEOUT": 10,
		"server.IDLE_TIMEOUT":  120,
		"database.DB_NAME":     "DB",
		"database.DB_USER":     "user",
		"database.DB_PASSWORD": "password",
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
	fmt.Println(banner)
	l := log.New(os.Stdout, "Config - ", log.LstdFlags)

	// set default config and load configuration from file
	defaultConfig := getDefaultConfiguration()
	configPaths, err := getConfigurationPaths()
	if err != nil {
		l.Printf("Error loading default config paths: %s\n", err)
		os.Exit(1)
	}
	configuration.SetDefaultConfiguration(defaultConfig)
	cf, err := configuration.LoadConfigurationFromFile(configFileName, configFileType, configPaths)
	if err != nil {
		l.Printf("Error reading configuration file: %s\n", err)
		os.Exit(1)
	}
	l.SetPrefix(fmt.Sprintf("%s - ", cf.API_NAME))

	// create db connection
	l.Println("Creating database connection")
	db, err := cf.Database.GetDatabaseConfiguration()
	if err != nil {
		l.Printf("Error reading database configuration: %s\n", err)
		os.Exit(1)
	}
	l.Printf("Configured database with: %s\n", db)

	// create the handlers
	l.Println("Creating handlers")

	// create a new serve mux and register handlers
	l.Println("Creating serve mux and registering handlers")
	sm := mux.NewRouter()

	// create a server
	s := http.Server{
		Addr:         cf.Server.BIND_ADDRESS,                // configure the bind address
		Handler:      sm,                                    // set the default handler
		ErrorLog:     l,                                     // set the logger for the server
		ReadTimeout:  cf.Server.READ_TIMEOUT * time.Second,  // max time to read request from the client
		WriteTimeout: cf.Server.WRITE_TIMEOUT * time.Second, // max time to write response to the client
		IdleTimeout:  cf.Server.IDLE_TIMEOUT * time.Second,  // max time for connections using TCP Keep-Alive
	}

	sc, err := cf.Server.GetServerConfiguration()
	if err != nil {
		l.Printf("Error reading server configuration: %s\n", err)
		os.Exit(1)
	}
	l.Printf("Configured server with: %s\n", sc)

	// start the server
	go func() {
		l.Printf("Starting server on: %s\n", cf.Server.BIND_ADDRESS)

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
	l.Println("Shutting down server gracefully")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	s.Shutdown(ctx)

}
