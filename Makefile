KEEL_API_DIR ?= ../keel-api
KEEL_API_SPEC = $(KEEL_API_DIR)/docs/public-artifacts/openapi.json
VENDORED_SPEC := api/openapi.json
OPENAPI_SPEC ?= $(VENDORED_SPEC)
OAPI_CODEGEN ?= go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen
GENERATED_CLIENT := pkg/keelclient/client.gen.go
SPEC_SOURCE := pkg/keelclient/SPEC_SOURCE
NORMALIZED_OPENAPI := .tmp/openapi.oapi-codegen.json
SHA256 := $(shell command -v sha256sum 2>/dev/null || echo shasum -a 256)

.PHONY: generate
generate:
	mkdir -p pkg/keelclient .tmp
	go run ./cmd/normalize-openapi -in $(OPENAPI_SPEC) -out $(NORMALIZED_OPENAPI)
	$(OAPI_CODEGEN) -config oapi-codegen.yaml -o $(GENERATED_CLIENT) $(NORMALIZED_OPENAPI)

# Redact prose from the spec committed at HEAD of a keel-api checkout,
# regenerate the client, and record both the source and vendored digests.
# While the source bytes are unchanged, a newer keel-api commit alone does
# not change the recorded commit.
.PHONY: regenerate
regenerate:
	@git -C $(KEEL_API_DIR) diff --quiet HEAD -- docs/public-artifacts/openapi.json || { echo '$(KEEL_API_SPEC) differs from its HEAD commit' >&2; exit 1; }
	mkdir -p $(dir $(VENDORED_SPEC))
	python3 scripts/redact_openapi.py --input $(KEEL_API_SPEC) --output $(VENDORED_SPEC)
	$(MAKE) generate OPENAPI_SPEC=$(VENDORED_SPEC)
	@commit=$$(git -C $(KEEL_API_DIR) rev-parse HEAD); \
	source_hash=$$($(SHA256) $(KEEL_API_SPEC) | cut -d ' ' -f 1); \
	if [ -f $(SPEC_SOURCE) ] && [ "$$source_hash" = "$$(sed -n 's/^source_sha256: //p' $(SPEC_SOURCE))" ]; then \
	  commit=$$(sed -n 's/^keel_api_commit: //p' $(SPEC_SOURCE)); \
	fi; \
	$(MAKE) spec-source OPENAPI_SPEC=$(VENDORED_SPEC) KEEL_API_COMMIT=$$commit KEEL_API_SOURCE_HASH=$$source_hash

# Write $(SPEC_SOURCE) for OPENAPI_SPEC, which came from keel-api at
# KEEL_API_COMMIT.
.PHONY: spec-source
spec-source:
	@test -n '$(KEEL_API_COMMIT)' || { echo 'KEEL_API_COMMIT is not set' >&2; exit 1; }
	@test -n '$(KEEL_API_SOURCE_HASH)' || { echo 'KEEL_API_SOURCE_HASH is not set' >&2; exit 1; }
	@{ \
	  echo '# Source of client.gen.go. Written by make regenerate; do not edit.'; \
	  echo 'keel_api_repository: keelapi/keel-api'; \
	  echo 'keel_api_commit: $(KEEL_API_COMMIT)'; \
	  echo 'spec_path: docs/public-artifacts/openapi.json'; \
	  echo 'source_sha256: $(KEEL_API_SOURCE_HASH)'; \
	  echo "spec_sha256: $$($(SHA256) $(OPENAPI_SPEC) | cut -d ' ' -f 1)"; \
	  echo 'redactor: scripts/redact_openapi.py'; \
	  echo "generator: $$(go list -m -f '{{.Path}} {{.Version}}' github.com/oapi-codegen/oapi-codegen/v2)"; \
	  echo 'generator_config: oapi-codegen.yaml'; \
	  echo "generator_flags: $$(awk '$$2 == "true" { sub(":", "", $$1); printf "%s%s", sep, $$1; sep = " " }' oapi-codegen.yaml)"; \
	  echo 'normalizer: cmd/normalize-openapi'; \
	} > $(SPEC_SOURCE)

# Regenerate from $(VENDORED_SPEC) and fail if the committed client or
# $(SPEC_SOURCE) differs from the result.
.PHONY: check-generated
check-generated:
	PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s scripts -p test_redact_openapi.py
	python3 scripts/check_redacted_spec.py $(VENDORED_SPEC)
	$(MAKE) generate OPENAPI_SPEC=$(VENDORED_SPEC)
	$(MAKE) spec-source OPENAPI_SPEC=$(VENDORED_SPEC) KEEL_API_COMMIT=$$(sed -n 's/^keel_api_commit: //p' $(SPEC_SOURCE)) KEEL_API_SOURCE_HASH=$$(sed -n 's/^source_sha256: //p' $(SPEC_SOURCE))
	@git diff --exit-code --stat -- $(GENERATED_CLIENT) $(SPEC_SOURCE) || { echo 'Generated files do not match $(VENDORED_SPEC): run make generate or make regenerate, and commit the result.' >&2; exit 1; }

README_EXAMPLE := .tmp/readme-example/main.go

# Compile the Go example under "## Basic Usage" in README.md.
.PHONY: check-readme
check-readme:
	mkdir -p $(dir $(README_EXAMPLE))
	awk '/^## Basic Usage/ {f = 1} f && /^```go$$/ {g = 1; next} g && /^```$$/ {exit} g' README.md | gofmt > $(README_EXAMPLE)
	@test -s $(README_EXAMPLE) || { echo 'README.md has no Go example under "## Basic Usage"' >&2; exit 1; }
	go vet ./$(README_EXAMPLE)
