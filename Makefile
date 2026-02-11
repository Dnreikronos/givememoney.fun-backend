.PHONY: build run test docker-up elk-up elk-down elk-logs elk-reset monitoring-up monitoring-down monitoring-logs monitoring-reset
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

monitoring-up:
	@docker-compose up -d prometheus grafana

monitoring-down:
	@docker-compose stop prometheus grafana
	@docker-compose rm -f prometheus grafana

monitoring-logs:
	@docker-compose logs -f prometheus grafana

monitoring-reset:
	@docker-compose stop prometheus grafana
	@docker-compose rm -f prometheus grafana
	@docker volume rm -f $$(docker volume ls -q | grep prometheus_data) 2>/dev/null || true
	@docker volume rm -f $$(docker volume ls -q | grep grafana_data) 2>/dev/null || true
