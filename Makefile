SRC_REID 		:=  $(wildcard reid/*.go)
SRC_CMD_COMMON 	:= $(wildcard cmd/common/*.go)

SRC_CMD_ENXML 	:= cmd/reid-enxml.go $(SRC_CMD_COMMON) $(SRC_REID)
SRC_CMD_CONVERT := cmd/reid-convert.go $(SRC_CMD_COMMON) $(SRC_REID)
SRC_CMD_SEARCH  := cmd/reid-search.go $(SRC_CMD_COMMON) $(SRC_REID)

# De-dup and sort
SRC_ALL := $(sort $(SRC_CMD_ENXML) $(SRC_CMD_CONVERT) $(SRC_CMD_SEARCH) $(SRC_CMD_COMMON) $(SRC_REID))

DEPS := .deps/kingpin.v2 .deps/docconv
COMMANDS := reid-enxml reid-convert reid-search

GO 		?= go
GOFMT 	?= gofmt

all: $(DEPS) $(COMMANDS)

reid-enxml: $(SRC_CMD_ENXML)
	$(GO) build $<

reid-convert: $(SRC_CMD_CONVERT)
	$(GO) build $<

reid-search: $(SRC_CMD_SEARCH)
	$(GO) build $<

.deps/kingpin.v2: .deps
	$(GO) get -v gopkg.in/alecthomas/kingpin.v2 && touch $@

.deps/docconv: .deps
	$(GO) get -v github.com/sajari/docconv && touch $@

.deps:
	@mkdir -p .deps

format: $(SRC_ALL)
	$(GOFMT) -w $(SRC_ALL)

clean:
	rm -f $(COMMANDS)

realclean: clean
	rm -rf .deps

.PHONY: format clean realclean
