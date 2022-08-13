###
#  general configuration of the makefile itself
###
SHELL := /bin/bash
.DEFAULT_GOAL := help

HELP_GENERATOR = mh
COMPOSE ?= docker compose

.PHONY: help netbox-launch netbox-data-import netbox-data-dump

help:
	@$(HELP_GENERATOR) -f $(MAKEFILE_LIST) $(target) 2>/dev/null || echo install mh from: https://github.com/oz123/mh/releases/tag/v1.0.1
ifndef target
	@type -P $(HELP_GENERATOR)>/dev/null && (echo ""; echo "Use \`make help target=foo\` to learn more about foo.")
endif


netbox-launch: ## start netbox with docker compose
	$(COMPOSE) up -d
					

netbox-data-import: ## import the data into postgresql
	$(COMPOSE) exec  postgres bash -c "apk add postgresql-client"
	$(COMPOSE) exec  postgres bash -c 'export PGPASSWORD=$(POSTGRES_PASSWORD); dropdb --if-exists -U netbox -h localhost netbox'
	$(COMPOSE) exec  postgres bash -c 'export PGPASSWORD=$(POSTGRES_PASSWORD); createdb -U netbox -h localhost netbox'
	$(COMPOSE) cp    netbox.sql postgres:/
	$(COMPOSE) exec  postgres bash -c 'export PGPASSWORD=$(POSTGRES_PASSWORD); psql -U netbox -h localhost -d netbox < netbox.sql'


netbox-data-dump:  ## dump sql data using sqldump.
	$(COMPOSE) exec postgres bash -c 'export PGPASSWORD=$(POSTGRES_PASSWORD); pg_dump -U netbox -h localhost -d netbox' > netbox.sql
