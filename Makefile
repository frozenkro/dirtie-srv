.PHONY: test

build:
	docker-compose up -d --build

run:
	docker-compose up -d mosquitto influxdb hub grafana postgres

dev:
	docker-compose up -d mosquitto influxdb grafana postgres

debug:
	docker-compose up -d mosquitto influxdb grafana postgres
	sleep 1
	dlv debug ./cmd/main.go

test:
	docker-compose up -d pg_test ix_test
	sleep 1
	go test ./... || docker-compose rm -fsv pg_test
	docker-compose rm -fsv pg_test ix_test
