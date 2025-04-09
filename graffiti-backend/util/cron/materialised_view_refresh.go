package cron

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robfig/cron/v3"
	"log"
	"time"
)

func ScheduleMaterializedViewRefresh(db *pgxpool.Pool) {
	c := cron.New(cron.WithLocation(time.FixedZone("Asia/Singapore", 8*3600)))
	_, err := c.AddFunc("0 3 * * *", func() { // Every day at 3AM
		log.Println("Refreshing materialized view via cron...")
		_, err := db.Exec(context.Background(), "REFRESH MATERIALIZED VIEW accepted_friendships_mv")
		if err != nil {
			log.Printf("Error refreshing materialized view: %v", err)
		} else {
			log.Println("View refreshed successfully.")
		}
	})
	if err != nil {
		log.Printf("Error scheduling cron job: %v", err)
		return
	}
	c.Start()
}
