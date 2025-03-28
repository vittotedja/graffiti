package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"github.com/vittotedja/graffiti/graffiti-backend/api"
	db "github.com/vittotedja/graffiti/graffiti-backend/db/sqlc"
	"github.com/vittotedja/graffiti/graffiti-backend/util"
)

var (
	dbSource string
)

func main() {
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

	ctx := context.Background()

	connPool, err := pgxpool.New(ctx, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	hub := db.NewHub(connPool)
	server := api.NewServer(hub)

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}

}
