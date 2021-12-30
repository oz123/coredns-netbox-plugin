SHELL := /bin/bash
VERSION ?= 1.8.6

.DEFAULT_GOAL := help

HELP_GENERATOR = mh

.PHONY: help
help:
	@$(HELP_GENERATOR) -f $(MAKEFILE_LIST) $(target) 2>/dev/null || echo install mh from: https://github.com/oz123/mh/releases/tag/v1.0.1
ifndef target
	@type -P $(HELP_GENERATOR)>/dev/null && (echo ""; echo "Use \`make help target=foo\` to learn more about foo.")
endif


coredns-fetch:
	 wget https://github.com/coredns/coredns/archive/refs/tags/v$(VERSION).zip -O coredns-v$(VERSION).zip


coredns-unzip:
	unzip coredns-v$(VERSION).zip


.PHONY: coredns-patch-go.mod
coredns-patch-go.mod:
	echo 'replace github.com/oz123/coredns-netbox-plugin =>' $(CURDIR) >> coredns-$(VERSION)/go.mod
	echo 'netbox:github.com/oz123/coredns-netbox-plugin' >> coredns-$(VERSION)/plugin.cfg

.PHONY: coredns-build
coredns-build:
	go get github.com/oz123/coredns-netbox-plugin
	go get github.com/coredns/coredns/plugin/etcd
	cd coredns-1.8.4/ && go get github.com/oz123/coredns-netbox-plugin
	make -C coredns-$(VERSION)/

.PHONY: coredns-run ## run the compiled version with plugin
coredns-run:
	./coredns-$(VERSION)/coredns -conf Corefile.example -p 5300

.PHONY: coredns-clean
coredns-clean:
	make -C coredns-$(VERSION) clean

.PHONY: test
test:  ## run the unit tests
	go test -v -failfast

