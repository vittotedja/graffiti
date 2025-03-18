postgres:
	docker run --name postgres -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret1234 -d postgres:alpine3.21

createdb:
	docker exec -it postgres createdb --username=root --owner=root graffiti

dropdb:
	docker exec -it postgres dropdb graffiti

sqlc:
	sqlc generate

server:
	go run main.go

.PHONY: postgres createdb dropdb sqlc server