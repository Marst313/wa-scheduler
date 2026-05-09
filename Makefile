.PHONY: *

run:
	-docker compose -f ./deploy/local/run/docker-compose.yml -p wa-scheduler down --remove-orphans
	docker compose -f ./deploy/local/run/docker-compose.yml -p wa-scheduler up --build --attach=server-scheduler