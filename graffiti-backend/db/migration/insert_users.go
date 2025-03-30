package main

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/vittotedja/graffiti/graffiti-backend/db/sqlc"
	"github.com/vittotedja/graffiti/graffiti-backend/util"
)

const totalUsers = 10000
const maxConcurrency = 10

func main() {
	ctx := context.Background()

	conn, err := pgxpool.New(ctx, "postgresql://root:secret1234@localhost:5432/graffiti?sslmode=disable")
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer conn.Close()

	hub := db.New(conn)

	log.Println("🚀 Inserting + onboarding + updating profiles...")
	start := time.Now()

	semaphore := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	rand.Seed(time.Now().UnixNano())

	for i := 0; i < totalUsers; i++ {
		wg.Add(1)
		semaphore <- struct{}{} // acquire

		go func(i int) {
			defer wg.Done()
			defer func() { <-semaphore }() // release

			arg := db.CreateUserParams{
				Username:       util.RandomUsername(),
				Fullname:       pgtype.Text{String: util.RandomFullname(), Valid: true},
				Email:          util.RandomEmail(),
				HashedPassword: "$2a$10$prehashedDummyPasswordForSeed",
			}

			user, err := hub.CreateUser(ctx, arg)
			if err != nil {
				log.Printf("user %d insert error: %v", i+1, err)
				return
			}

			// Mark as onboarded
			if err := hub.FinishOnboarding(ctx, user.ID); err != nil {
				log.Printf("user %d onboarding error: %v", i+1, err)
				return
			}

			// Randomize bio and profile_picture (50% chance each)
			var bio pgtype.Text
			if rand.Intn(2) == 0 {
				bio = pgtype.Text{String: util.RandomBio(), Valid: true}
			} else {
				bio = pgtype.Text{Valid: false}
			}

			var pfp pgtype.Text
			if rand.Intn(2) == 0 {
				pfp = pgtype.Text{String: util.RandomProfilePictureURL(), Valid: true}
			} else {
				pfp = pgtype.Text{Valid: false}
			}

			profileArg := db.UpdateProfileParams{
				ID:              user.ID,
				ProfilePicture:  pfp,
				Bio:             bio,
				BackgroundImage: pgtype.Text{Valid: false}, // skip for now
			}

			if _, err := hub.UpdateProfile(ctx, profileArg); err != nil {
				log.Printf("user %d profile update error: %v", i+1, err)
				return
			}

			if (i+1)%1000 == 0 {
				log.Printf("%d users processed", i+1)
			}
		}(i)
	}

	wg.Wait()
	log.Printf("Done seeding %d users in %v", totalUsers, time.Since(start))
}
