.PHONY: test

test:
	docker-compose up -d pg_test
	sleep 2
	go test ./...
	docker-compose rm -fsv pg_test
