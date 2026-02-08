.PHONY: build run test docker-up elk-up elk-down elk-logs elk-reset
build:
	@go build -o givememoney.fun-backend cmd/main.go

docker-up:
	@docker-compose up -d

run:
	@go run cmd/main.go

test:
	@go test ./...

elk-up:
	@docker-compose up -d elasticsearch logstash kibana filebeat

elk-down:
	@docker-compose stop elasticsearch logstash kibana filebeat
	@docker-compose rm -f elasticsearch logstash kibana filebeat

elk-logs:
	@docker-compose logs -f elasticsearch logstash kibana filebeat

elk-reset:
	@docker-compose stop elasticsearch logstash kibana filebeat
	@docker-compose rm -f elasticsearch logstash kibana filebeat
	@docker volume rm -f $$(docker volume ls -q | grep elasticsearch_data) 2>/dev/null || true
	@docker volume rm -f $$(docker volume ls -q | grep filebeat_data) 2>/dev/null || true
