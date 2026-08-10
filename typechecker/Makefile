.PHONY: build deploy git-status test pbt continuouspbt slowtest fmt clean

ifeq ($(OS),Windows_NT)
    BINEXT := .exe
else
    BINEXT :=
endif


DEPLOYPATH ?= Z:/Software/TypeChecker/
EXENAME    ?= suneidotypes.exe
BIN        := bin/suneidotypes$(BINEXT)
LINUXBIN   ?= suneidotypes_linux_amd64

help: ## prints this message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "\033[1;33m%-30s\033[0m %s\n", $$1, $$2}'

# dev build - no version stamping, main.builtDate stays "dev build"
build: fmt ## build type checker binary
	mkdir -p bin/
	go build -o $(BIN) ./cmd/suneidotypes
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/$(LINUXBIN) ./cmd/suneidotypes

# `=` (not `:=`) so shells run only when deploy is the target
BUILT_DATE = $(shell date +'%b %d %Y %H:%M' 2>/dev/null || powershell -NoProfile -Command "Get-Date -Format 'MMM dd yyyy HH:mm'")
GIT_COMMIT = $(shell git rev-parse --short=7 HEAD)
LDFLAGS    = -s -w -X 'main.builtDate=$(BUILT_DATE) $(GIT_COMMIT)'

# TODO: uncomment when no longer in active development
deploy: test # git-status ## build, test, and deploy to Z
	GOGC=off go build -ldflags="$(LDFLAGS)"		 -o $(BIN) ./cmd/suneidotypes
	CGO_ENABLED=0 GOGC=off GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)"	 	 -o $(LINUXBIN) ./cmd/suneidotypes
	@echo -e "\n"
	mv $(BIN) $(DEPLOYPATH)$(EXENAME)
	ls -lh $(DEPLOYPATH)$(EXENAME)
	@echo -e "\n"
	mv $(LINUXBIN) $(DEPLOYPATH)$(LINUXBIN)
	ls -lh $(DEPLOYPATH)$(LINUXBIN)
	@echo -e "\n$(BUILT_DATE) $(GIT_COMMIT)"

git-status:
	@test -z "$(shell git status --porcelain)"

# property-based tests are named TestProp* and live in ./typeinfer and
# ./typeinfer/typealgebra. the rapid iteration count is scoped to these
# packages because -rapid.checks is only a valid flag for test binaries that
# import rapid. override on the CLI, e.g. `make pbt PBT_ITERS=5000`.
# NOTE: each package gets its own go test invocation (see pbt) - `go test -v`
# only streams output live for a single package; given several it buffers
# each package's output until that package finishes, which makes a long
# property run look hung. fast package first for quick signal.
PBT_PKG     := ./typeinfer/typealgebra/ ./typeinfer/internal/engine/
PBT_PATTERN := TestProp
PBT_ITERS   ?= 500
SLOW_ITERS  ?= 5000
# many rapid properties run the whole pipeline per check, so a high-iteration
# run can exceed go test's default 10m timeout. override with e.g. PBT_TIMEOUT=1h.
PBT_TIMEOUT ?= 30m

test: ## run all go tests (regular) + property-based tests
	go test ./... -v -count=1 -skip '$(PBT_PATTERN)'
	@for pkg in $(PBT_PKG); do \
		go test $$pkg -run '$(PBT_PATTERN)' -count=1 -v -timeout $(PBT_TIMEOUT) -rapid.checks=$(PBT_ITERS) || exit 1; \
	done

pbt: ## run only the property-based tests (rapid, PBT_ITERS iters)
	@for pkg in $(PBT_PKG); do \
		go test $$pkg -run '$(PBT_PATTERN)' -count=1 -v -timeout $(PBT_TIMEOUT) -rapid.checks=$(PBT_ITERS) || exit 1; \
	done

# keep running the property tests round after round until one fails. rapid
# time-seeds when -rapid.seed is unset (like `make pbt`), so every round
# explores fresh inputs. stops on the first failing round, leaving rapid's
# .fail repro file + the `-run/-rapid.failfile` line it prints. CPBT_ROUNDS=0
# loops forever (Ctrl-C to stop); set e.g. `make continuouspbt CPBT_ROUNDS=50`.
CPBT_ROUNDS ?= 0
CPBT_ITERS  ?= 1000

continuouspbt: ## loop pbt until a failure (CPBT_ROUNDS=0 = forever, Ctrl-C to stop)
	@round=0; \
	while true; do \
		round=$$((round+1)); \
		echo "=== continuouspbt round $$round ($(CPBT_ITERS) checks/pkg) ==="; \
		for pkg in $(PBT_PKG); do \
			go test $$pkg -run '$(PBT_PATTERN)' -count=1 -v -timeout $(PBT_TIMEOUT) -rapid.checks=$(CPBT_ITERS) \
				|| { echo "FAILED on round $$round (pkg $$pkg) - see rapid .fail repro above"; exit 1; }; \
		done; \
		if [ "$(CPBT_ROUNDS)" -ne 0 ] && [ $$round -ge "$(CPBT_ROUNDS)" ]; then \
			echo "completed $(CPBT_ROUNDS) rounds, no failures"; break; \
		fi; \
	done

slowtest: ## run all tests + property-based tests with many iterations
	go test ./... -v -count=1 -skip '$(PBT_PATTERN)'
	@for pkg in $(PBT_PKG); do \
		go test $$pkg -run '$(PBT_PATTERN)' -count=1 -v -timeout $(PBT_TIMEOUT) -rapid.checks=$(SLOW_ITERS) || exit 1; \
	done

fmt: ## run go formatter
	go fmt ./...

lint: ## run go ci linter
	golangci-lint run ./...

clean:
	rm -rf ./bin/

