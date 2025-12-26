.PHONY: up down logs demo

up:
	docker-compose -f docker-compose.yaml up --build

down:
	docker-compose -f docker-compose.yaml down

logs:
	docker-compose -f docker-compose.yaml logs -f --tail=200

demo:
	bash scripts/demo.sh
