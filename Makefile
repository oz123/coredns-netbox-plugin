SHELL := /bin/bash
VERSION ?= 1.12.0

.DEFAULT_GOAL := help

HELP_GENERATOR = mh

.PHONY: help
help:
	@$(HELP_GENERATOR) -f $(MAKEFILE_LIST) $(target) 2>/dev/null || echo install mh from: https://github.com/oz123/mh/releases/
ifndef target
	@type -P $(HELP_GENERATOR)>/dev/null && (echo ""; echo "Use \`make help target=foo\` to learn more about foo.")
endif


coredns-fetch: ## download coredns to local directory
	 wget https://github.com/coredns/coredns/archive/refs/tags/v$(VERSION).zip -O coredns-v$(VERSION).zip


coredns-unzip: ## unzip coredns distribution
	unzip coredns-v$(VERSION).zip


.PHONY: coredns-patch-go.mod
coredns-patch-go.mod:  ## patch coredns to local compile
	grep netbox-plugin coredns-$(VERSION)/go.mod || echo 'replace github.com/oz123/coredns-netbox-plugin =>' $(CURDIR) >> coredns-$(VERSION)/go.mod

.PHONY: coredns-add-rest-plugin
coredns-add-rest-plugin:  ## patch coredns to local compile
	grep netbox-plugin coredns-$(VERSION)/go.mod || sed '/hosts:hosts/i netbox:github.com/oz123/coredns-netbox-plugin' plugin.cfg coredns-$(VERSION)/plugin.cfg

.PHONY: coredns-build
coredns-build:  ## build local coredns with the plugin installed
	make -C coredns-$(VERSION)/

.PHONY: coredns-run ## run the compiled version with plugin
coredns-run:  ## run patched coredns
	./coredns-$(VERSION)/coredns -conf Corefile.example -p 5300

.PHONY: coredns-clean
coredns-clean:
	make -C coredns-$(VERSION) clean

.PHONY: test
test:  ## run the unit tests
	go test -v -failfast

test-api:
	cd testing; python3 api.py
