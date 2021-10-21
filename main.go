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
	"github.com/spf13/viper"
)

func main() {
	// set defaults and read config file
	viper.SetConfigName("config") // name of config file (without extension)
	// viper.SetConfigType("yaml")     // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("/etc/app") // path to look for the config file in
	viper.AddConfigPath("$HOME/.app")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// config file was found but another error was produced
			pe := fmt.Errorf("error reading config file: %w", err)
			fmt.Println(pe)
			os.Exit(1)
		}
	}
	viper.SetDefault("BIND_ADDRESS", ":9090")
	viper.SetDefault("API_NAME", "API")
	viper.SetDefault("READ_TIMEOUT", 5)
	viper.SetDefault("WRITE_TIMEOUT", 10)
	viper.SetDefault("IDLE_TIMEOUT", 120)

	l := log.New(os.Stdout, fmt.Sprintf("%s: ", viper.GetString("API_NAME")), log.LstdFlags)

	// create the handlers
	l.Println("Creating handlers")

	// create a new serve mux and register handlers
	l.Println("Creating serve mux and registering handlers")
	sm := mux.NewRouter()

	// create a server
	s := http.Server{
		Addr:         viper.GetString("BIND_ADDRESS"),                  // configure the bind address
		Handler:      sm,                                               // set the default handler
		ErrorLog:     l,                                                // set the logger for the server
		ReadTimeout:  viper.GetDuration("READ_TIMEOUT") * time.Second,  // max time to read request from the client
		WriteTimeout: viper.GetDuration("WRITE_TIMEOUT") * time.Second, // max time to write response to the client
		IdleTimeout:  viper.GetDuration("IDLE_TIMEOUT") * time.Second,  // max time for connections using TCP Keep-Alive
	}

	// TODO: Make this look better
	l.Printf("Created server: %+v\n", s)

	// start the server
	go func() {
		l.Printf("Starting server on %s\n", viper.GetString("BIND_ADDRESS"))

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
