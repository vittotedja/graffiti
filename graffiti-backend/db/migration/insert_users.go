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
const maxFriendsPerUser = 10 // Adjust as needed

func main() {
	ctx := context.Background()

	config, err := util.LoadConfig("../../.")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	conn, err := pgxpool.New(ctx, config.DBSource)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer conn.Close()

	hub := db.New(conn)

	log.Println("🚀 Inserting users and setting up profiles...")
	start := time.Now()

	semaphore := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup
	rand.Seed(time.Now().UnixNano())

	// Pre-allocate users for friendship processing
	users := make([]db.User, totalUsers)

	for i := 0; i < totalUsers; i++ {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(i int) {
			defer wg.Done()
			defer func() { <-semaphore }()

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
			users[i] = user

			_ = hub.FinishOnboarding(ctx, user.ID)

			profileArg := db.UpdateProfileParams{
				ID:              user.ID,
				ProfilePicture:  pgtype.Text{String: util.RandomProfilePictureURL(), Valid: true},
				Bio:             pgtype.Text{String: util.RandomBio(), Valid: true},
				BackgroundImage: pgtype.Text{Valid: false},
			}
			if _, err := hub.UpdateProfile(ctx, profileArg); err != nil {
				log.Printf("user %d profile update error: %v", i+1, err)
			}
			if (i+1)%1000 == 0 {
				log.Printf("%d users processed", i+1)
			}
		}(i)
	}

	wg.Wait()
	log.Println("✅ Users seeded.")

	log.Println("👥 Creating friendships...")

	var friendshipWg sync.WaitGroup
	friendshipSemaphore := make(chan struct{}, maxConcurrency)

	for i := 0; i < totalUsers; i++ {
		user := users[i]
		friendCount := rand.Intn(maxFriendsPerUser) + 1

		for j := 0; j < friendCount; j++ {
			friendIndex := rand.Intn(totalUsers)
			if friendIndex == i {
				continue // avoid self friendship
			}
			friend := users[friendIndex]

			friendshipWg.Add(1)
			friendshipSemaphore <- struct{}{}

			go func(fromUser, toUser pgtype.UUID) {
				defer friendshipWg.Done()
				defer func() { <-friendshipSemaphore }()

				// Insert one-way friendship
				_, err := hub.CreateFriendship(ctx, db.CreateFriendshipParams{
					FromUser: fromUser,
					ToUser:   toUser,
					Status: db.NullStatus{
						Status: "friends",
						Valid:  true,
					},
				})
				if err != nil {
					log.Printf("friendship error (%s -> %s): %v", fromUser, toUser, err)
				}
			}(user.ID, friend.ID)
		}
	}

	friendshipWg.Wait()

	// Refresh materialized view
	log.Println("🔄 Refreshing materialized view...")
	_, err = conn.Exec(ctx, "REFRESH MATERIALIZED VIEW accepted_friendships_mv")
	if err != nil {
		log.Fatalf("error refreshing materialized view: %v", err)
	}

	log.Printf("✅ Done in %v", time.Since(start))
}
