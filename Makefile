OPENAPI_CFG=openapi/oapi-codegen-config.yaml

generate-api:
	oapi-codegen --config=$(OPENAPI_CFG) openapi/orbitum.yml

run-http:
	go run ./cmd/orbitum

tidy:
	go mod tidy