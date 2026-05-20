OPENAPI_SPEC ?= ../keel-api/docs/public-artifacts/openapi.json
OAPI_CODEGEN ?= go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen
GENERATED_CLIENT := pkg/keelclient/client.gen.go
NORMALIZED_OPENAPI := .tmp/openapi.oapi-codegen.json

.PHONY: generate
generate:
	mkdir -p pkg/keelclient .tmp
	go run ./cmd/normalize-openapi -in $(OPENAPI_SPEC) -out $(NORMALIZED_OPENAPI)
	$(OAPI_CODEGEN) -config oapi-codegen.yaml -o $(GENERATED_CLIENT) $(NORMALIZED_OPENAPI)
