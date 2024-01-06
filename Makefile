.PHONY: prep
prep:
	docker-compose build

.PHONY: start
start_local:
	docker-compose --env-file .env.local up
