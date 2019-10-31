.PHONY: run
run:
	docker-compose up

.PHONY: exec
exec:
	docker-compose exec roach1 ./cockroach sql --insecure -d="bank"

.PHONY: bootstrap
bootstrap:
	cat bootstrap.sql | docker-compose exec -T roach1 ./cockroach sql --insecure --echo-sql

.PHONY: test
test: bootstrap
	cd goapp && go run main.go

.PHONY: rm
rm:
	docker-compose down; docker-compose rm
