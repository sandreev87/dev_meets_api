.PHONY: prep
prep:
	docker-compose -f docker-compose.local.yml build

.PHONY: start
start_local:
	docker-compose --env-file .env.local -f docker-compose.local.yml up
