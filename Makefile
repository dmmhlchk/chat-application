PHONY: communication-api-service

communication-api-service:
	go build -o ./builds/communication-api-service  -v ./cmd/communication-api-service


PHONY: identity-api-service
identity-api-service:
	go build -o ./builds/identity-api-service  -v ./cmd/identity-api-service