# Setting up standard path variables similar to autoconf
# The defaults are taken based on
# https://www.gnu.org/prep/standards/html_node/Directory-Variables.html
# and
# https://fedoraproject.org/wiki/Packaging:RPMMacros?rd=Packaging/RPMMacros

PREFIX ?= /usr/local

BASE_PREFIX = $(PREFIX)
ifeq ($(PREFIX), /usr)
BASE_PREFIX = ""
endif

EXEC_PREFIX ?= $(PREFIX)

BINDIR ?= $(EXEC_PREFIX)/bin
SBINDIR ?= $(EXEC_PREFIX)/sbin

DATADIR ?= $(PREFIX)/share
LOCALSTATEDIR ?= $(BASE_PREFIX)/var/lib
LOGDIR ?= $(BASE_PREFIX)/var/log

SYSCONFDIR ?= $(BASE_PREFIX)/etc
RUNDIR ?= $(BASE_PREFIX)/var/run


EXPORTER = gluster-exporter

BUILDDIR = build

EXPORTER_BIN = $(EXPORTER)
EXPORTER_BUILD = $(BUILDDIR)/$(EXPORTER_BIN)
EXPORTER_INSTALL = $(DESTDIR)$(SBINDIR)/$(EXPORTER_BIN)
EXPORTER_SERVICE_BUILD = $(BUILDDIR)/$(EXPORTER).service
EXPORTER_SERVICE_INSTALL = $(DESTDIR)/usr/lib/systemd/system/$(EXPORTER).service
EXPORTER_CONF_INSTALL = $(DESTDIR)$(SYSCONFDIR)/$(EXPORTER)

GD2STATEDIR = $(LOCALSTATEDIR)/glusterd2
GD1STATEDIR = $(LOCALSTATEDIR)/glusterd
EXPORTER_LOGDIR = $(LOGDIR)/$(EXPORTER)
EXPORTER_RUNDIR = $(RUNDIR)/$(EXPORTER)

GLUSTER_MGMT ?= "glusterd"

DEPENV ?=

FASTBUILD ?= yes

.PHONY: all build binaries check check-go check-reqs install vendor-update vendor-install verify release check-protoc $(EXPORTER_BIN) $(EXPORTER_BUILD) test dist dist-vendor gen-service gen-version metrics-docgen

all: build

build: check-go check-reqs vendor-install $(EXPORTER_BIN)
check: check-go check-reqs

check-go:
	@./scripts/check-go.sh
	@echo

check-reqs:
	@./scripts/check-reqs.sh
	@echo

$(EXPORTER_BIN): $(EXPORTER_BUILD) gen-service
$(EXPORTER_BUILD):
	FASTBUILD=$(FASTBUILD) BASE_PREFIX=$(BASE_PREFIX) GD1STATEDIR=$(GD1STATEDIR) \
		CONFFILE=${SYSCONFDIR}/gluster-exporter/gluster-exporter.toml \
		GD2STATEDIR=$(GD2STATEDIR) ./scripts/build.sh $(EXPORTER)
	@echo

install:
	install -D $(EXPORTER_BUILD) $(EXPORTER_INSTALL)
	install -D -m 0644 $(EXPORTER_SERVICE_BUILD) $(EXPORTER_SERVICE_INSTALL)
	install -D -m 0600 ./extras/conf/gluster-exporter.toml.sample $(EXPORTER_CONF_INSTALL)/gluster-exporter.toml
	@echo

vendor-update:
	@echo Updating vendored packages
	@$(DEPENV) dep ensure -v -update
	@echo

vendor-install:
	@echo Installing vendored packages
	@$(DEPENV) dep ensure -v -vendor-only
	@echo

test: check-reqs
	@./scripts/pre-commit.sh
	@./scripts/gometalinter-tests.sh
	@echo

release: build
	@./scripts/release.sh

dist: gen-version
	@DISTDIR=$(DISTDIR) SIGN=$(SIGN) ./scripts/dist.sh
	@rm -f ./VERSION ./GIT_SHA_FULL

dist-vendor: vendor-install gen-version
	@VENDOR=yes DISTDIR=$(DISTDIR) SIGN=$(SIGN) ./scripts/dist.sh
	@rm -f ./VERSION ./GIT_SHA_FULL

gen-service:
	SBINDIR=$(SBINDIR) SYSCONFDIR=$(SYSCONFDIR) GLUSTER_MGMT=$(GLUSTER_MGMT) \
		./scripts/gen-service.sh

gen-version:
	@git describe --tags --always --match "v[0-9]*" > ./VERSION
	@git rev-parse HEAD > ./GIT_SHA_FULL

metrics-docgen: $(EXPORTER_BIN)
	mkdir -p docs
	./build/gluster-exporter --docgen > docs/metrics.adoc
