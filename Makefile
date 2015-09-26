PROJ=github.com/mischief/glenda
PROGS=glenda glendactl
KITES=define feed geoip irc markov qdb

TOPDIR=$(PWD)
BIN=$(TOPDIR)/bin

all:	bin

bin:	$(PROGS:%=bin/%) $(KITES:%=bin/glenda-%)

bin/%:	$(TOPDIR)/cmd/%
	@echo building $*...
	@CGO_ENABLED=0 go build -ldflags "-s" -installsuffix nocgo -o $(BIN)/$* $(PROJ)/cmd/$*

#containers: bin glenda.docker $(KITES:%=glenda-%.docker)
containers: bin
	docker build -t mischief/glenda:latest .

#%.docker: $(TOPDIR)/cmd/%/Dockerfile
#	docker build -t $*:latest -f $< .

clean:
	rm -f bin/glenda*

.PHONY: bin containers

