package main

import (
	"context"
	"github.com/vittotedja/graffiti/graffiti-backend/util/logger"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"
	"github.com/vittotedja/graffiti/graffiti-backend/api"
	"github.com/vittotedja/graffiti/graffiti-backend/util"
)

var (
	dbSource string
)

func main() {
	if err := logger.Setup(); err != nil {
		log.Fatal("cannot setup logger:", err)
	}
	// r := gin.Default()
	// r.GET("/", func(c *gin.Context) {
	// 	c.JSON(http.StatusOK, gin.H{
	// 		"message": "Hello World",
	// 	})
	// })

	// r.Run(":8080")
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	// switch config.Env {
	// case "devlocal":
	//     dbSource = config.DBSourceLocal
	// case "devdocker":
	//     dbSource = config.DBSourceDocker
	// case "production":
	//     dbSource = "postgresql://<RDS_USER>:<RDS_PASSWORD>@<RDS_ENDPOINT>:5432/graffiti?sslmode=require"
	// default:
	//     fmt.Println("Unknown environment, using default database source")
	//     dbSource = config.DBSourceLocal
	// }

	server := api.NewServer(config)

	go func() {
		log.Printf("Starting server on port %s...", config.ServerAddress)
		if err := server.Start(); err != nil {
			log.Fatalf("Server encountered an error: %v", err)
		}
	}()

	// Set up channel to listen for OS signals (e.g., SIGINT, SIGTERM)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal
	<-stop
	log.Println("Received shutdown signal, shutting down gracefully...")

	// Gracefully shut down the server
	if err := server.Shutdown(); err != nil {
		log.Fatalf("Graceful shutdown failed: %v", err)
	}

	log.Println("Server shut down successfully")

}
