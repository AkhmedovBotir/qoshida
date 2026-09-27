.PHONY: run tidy migrate create-admin import-regions import-categories import-activity-types

run:
	go run ./cmd/api

tidy:
	go mod tidy

migrate:
	go run ./cmd/migrate

create-admin:
	go run ./cmd/create-admin

import-regions:
	go run ./cmd/import-regions

import-categories:
	go run ./cmd/import-categories

import-activity-types:
	go run ./cmd/import-activity-types
