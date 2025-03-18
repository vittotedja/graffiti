postgres:
	docker run --name postgres -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret1234 -d postgres:17

createdb:
	docker exec -it postgres createdb --username=root --owner=root graffiti

dropdb:
	docker exec -it postgres dropdb graffiti

sqlc:
	sqlc generate

migrateup:
	migrate -path db/migration -database "postgresql://root:secret1234@localhost:5432/graffiti?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://root:secret1234@localhost:5432/graffiti?sslmode=disable" -verbose down

test:
	go test -v -cover ./...

server:
	go run main.go

.PHONY: postgres createdb dropdb sqlc test server migrateup migratedown

