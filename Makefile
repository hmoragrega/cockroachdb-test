.PHONY: run
run:
	docker-compose up --force-recreate

.PHONY: sql
sql:
	docker-compose exec roach1 ./cockroach sql --insecure -d="bank"

.PHONY: bootstrap
bootstrap:
	cat bootstrap.sql | docker-compose exec -T roach1 ./cockroach sql --insecure --echo-sql

.PHONY: test
test: bootstrap
	cd goapp && go run -mod vendor main.go

.PHONY: stress
stress: bootstrap
	cd goapp && go build -mod vendor main.go && cd .. && ./stress.sh

.PHONY: rm
rm:
	docker-compose down; docker-compose rm
