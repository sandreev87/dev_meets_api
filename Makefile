.PHONY: prep
prep:
	docker-compose build --no-cache

.PHONY: start
start_local:
	docker-compose --env-file .env.local up
