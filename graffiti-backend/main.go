package main

import (
	logutil "github.com/vittotedja/graffiti/graffiti-backend/util/logger"
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
	if err := logutil.Setup(); err != nil {
		log.Fatal("cannot setup logger:", err)
	}

	logger := logutil.Global()
	logger.Info("Logger initialized successfully")

	// r := gin.Default()
	// r.GET("/", func(c *gin.Context) {
	// 	c.JSON(http.StatusOK, gin.H{
	// 		"message": "Hello World",
	// 	})
	// })

	// r.Run(":8080")
	config, err := util.LoadConfig(".")
	if err != nil {
		logger.Fatal("cannot load config:", err)
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
		logger.Info("Starting server in %s environment on %s", config.Env, config.ServerAddress)
		if err := server.Start(); err != nil {
			logger.Errorf("Server encountered an error: %v", err)
		}
	}()

	// Set up channel to listen for OS signals (e.g., SIGINT, SIGTERM)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal
	<-stop
	logger.Info("Received shutdown signal, shutting down gracefully...")

	// Gracefully shut down the server
	if err := server.Shutdown(); err != nil {
		logger.Fatal("Graceful shutdown failed: %v", err)
	}

	logger.Info("Server shut down successfully")

}
