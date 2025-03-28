# Variables
NETWORK_NAME := graffiti_dev_network
POSTGRES_CONTAINER := postgres
GRAFFITI_CONTAINER := graffiti

# Create new network for backend and postgres container to communicate
create-network:
	docker network create $(NETWORK_NAME)

# If you don't want to rebuild ur old postgres database, just use this (make connect-postgres-network)
connect-postgres-network:
	docker network connect $(NETWORK_NAME) $(POSTGRES_CONTAINER)

postgres:
	docker run --name $(POSTGRES_CONTAINER) --network $(NETWORK_NAME) -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret1234 -d postgres:17

createdb:
	docker exec -it $(POSTGRES_CONTAINER) createdb --username=root --owner=root graffiti

dropdb:
	docker exec -it $(POSTGRES_CONTAINER) dropdb graffiti
	
sqlc:
	sqlc generate

migrateup:
	migrate -path db/migration -database "postgresql://root:secret1234@localhost:5432/graffiti?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://root:secret1234@localhost:5432/graffiti?sslmode=disable" -verbose down

test:
	go test -v -cover ./...

# For local machine development
server:
	go run main.go

# Remove existing backend image and rebuild. If you don't have backend image yet, just run the third line on ur terminal
rebuild-server-docker:
	docker rm -f $(GRAFFITI_CONTAINER) && \
    docker rmi graffiti:latest && \
    docker build -t graffiti:latest .

# Run the image of backend in the same network as postgres
run-server-docker:
	docker run --name $(GRAFFITI_CONTAINER) --network $(NETWORK_NAME) -p 8080:8080 -e ENV=devdocker -e DB_SOURCE="postgresql://root:secret1234@$(POSTGRES_CONTAINER):5432/graffiti?sslmode=disable" -d graffiti:latest

.PHONY: create-network connect-postgres-network postgres createdb dropdb sqlc migrateup migratedown test server rebuild-server-docker run-server-docker

