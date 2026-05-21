.PHONY: *

test:
	docker compose -p wa-scheduler -f ./deploy/local/test/docker-compose.yml down --remove-orphans
	docker compose -p wa-scheduler -f ./deploy/local/test/docker-compose.yml up --exit-code-from=test --attach=test

run:
	-docker compose -f ./deploy/local/run/docker-compose.yml -p wa-scheduler down --remove-orphans
	docker compose -f ./deploy/local/run/docker-compose.yml -p wa-scheduler up --build --attach=server-scheduler