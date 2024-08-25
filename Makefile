up:
	docker compose up -d

down:
	docker compose down

start:
	docker compose start

stop:
	docker compose stop

in:
	docker compose exec ddlparse bash

build:
	docker compose build --no-cache

test:
	go test