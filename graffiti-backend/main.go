package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/vittotedja/graffiti/graffiti-backend/db/sqlc"
	"github.com/vittotedja/graffiti/graffiti-backend/api"
	_ "github.com/lib/pq"
)

const (
	dbDriver = "postgres"
	dbSource = "postgresql://root:secret1234@localhost:5432/graffiti?sslmode=disable"
	serverAddress = "0.0.0.0:8080"
)

func main() {
	// r := gin.Default()
	// r.GET("/", func(c *gin.Context) {
	// 	c.JSON(http.StatusOK, gin.H{
	// 		"message": "Hello World",
	// 	})
	// })

	// r.Run(":8080")
	ctx := context.Background()

	connPool, err := pgxpool.New(ctx, dbSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	hub := db.NewHub(connPool)
	server := api.NewServer(hub)

	err = server.Start(serverAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}

}
