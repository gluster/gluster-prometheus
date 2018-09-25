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

GD2STATEDIR = $(LOCALSTATEDIR)/glusterd2
GD1STATEDIR = $(LOCALSTATEDIR)/glusterd
EXPORTER_LOGDIR = $(LOGDIR)/$(EXPORTER)
EXPORTER_RUNDIR = $(RUNDIR)/$(EXPORTER)

GLUSTER_MGMT ?= "glusterd1"

DEPENV ?=

FASTBUILD ?= yes

.PHONY: all build binaries check check-go check-reqs install vendor-update vendor-install verify release check-protoc $(EXPORTER_BIN) $(EXPORTER_BUILD) test dist dist-vendor gen-service

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
		GD2STATEDIR=$(GD2STATEDIR) ./scripts/build.sh $(EXPORTER)
	@echo

install:
	install -D $(EXPORTER_BUILD) $(EXPORTER_INSTALL)
	install -D -m 0644 $(EXPORTER_SERVICE_BUILD) $(EXPORTER_SERVICE_INSTALL)
	@echo

vendor-update:
	@echo Updating vendored packages
	@$(DEPENV) dep ensure -update
	@echo

vendor-install:
	@echo Installing vendored packages
	@$(DEPENV) dep ensure -vendor-only
	@echo

test: check-reqs
	@./test.sh $(TESTOPTIONS)

release: build
	@./scripts/release.sh

dist:
	@DISTDIR=$(DISTDIR) SIGN=$(SIGN) ./scripts/dist.sh

dist-vendor: vendor-install
	@VENDOR=yes DISTDIR=$(DISTDIR) SIGN=$(SIGN) ./scripts/dist.sh

gen-service:
	SBINDIR=$(SBINDIR) GLUSTER_MGMT=$(GLUSTER_MGMT) ./scripts/gen-service.sh
