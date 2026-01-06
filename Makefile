filter-openapi:
	pnpx openapi-format http://localhost:8080/openapi.yaml -o dist/openapi-formatted.yaml --filterFile ./openapi-filters.yaml

downgrade-openapi:
	pnpx @apiture/openapi-down-convert -i dist/openapi-formatted.yaml -o dist/openapi-downgraded.yaml

generate-sdk: filter-openapi downgrade-openapi
	oapi-codegen -config cfg.yaml ./dist/openapi-downgraded.yaml

.PHONY: filter-openapi downgrade-openapi generate-sdk