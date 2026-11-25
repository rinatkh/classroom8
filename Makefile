SHELL := /bin/bash
.DEFAULT_GOAL := help
GO ?= go
UNIT_PACKAGES := ./internal/... ./cmd/benchreport
BENCHTIME ?= 100ms
BENCHCOUNT ?= 3
BENCHCPU ?= 1
COVERAGE_FILE ?= coverage.out

fmt: ## КАЧЕСТВО | Применить gofmt
	gofmt -w $$(find . -name '*.go' -not -path './bin/*')
fmt-check: ## КАЧЕСТВО | Проверить форматирование без изменения файлов
	@files="$$(gofmt -l $$(find . -name '*.go' -not -path './bin/*'))"; if [ -n "$$files" ]; then echo "$$files"; exit 1; fi
vet: ## КАЧЕСТВО | Статический анализ go vet
	$(GO) vet ./...
compile: ## КАЧЕСТВО | Компиляция всех пакетов и тестов без выполнения заданий
	$(GO) test -run '^$$' ./...
test-unit: ## ТЕСТЫ | Все модульные тесты internal
	$(GO) test $(UNIT_PACKAGES)
test-race: ## ТЕСТЫ | Все тесты с race detector
	$(GO) test -race ./...
coverage: ## КАЧЕСТВО | Statements coverage всех internal-пакетов
	$(GO) test $(UNIT_PACKAGES) -coverpkg=./internal/... -covermode=atomic -coverprofile=$(COVERAGE_FILE)
	$(GO) tool cover -func=$(COVERAGE_FILE)
clean: ## СБОРКА | Удалить bin и coverage.out
	rm -rf bin $(COVERAGE_FILE)

CMDS := 01_linkedlist 02_stack 03_queue 04_buffer 05_ringbuffer 06_hashtable 07_visualizations
BUILD_CMDS := $(CMDS) benchreport
BENCH_GROUPS := list stack queue hash buffer memory
BENCH_TARGETS := $(addprefix bench-,$(BENCH_GROUPS))
BENCHVIEW ?= compact
BENCHSIZE ?= 0
BENCHOP ?=
BENCHDIR ?= bin/benchmarks
COVERAGE_THRESHOLD ?= 90.0
PACKAGE_FILE ?= bin/classroom8-$(shell $(GO) env GOOS)-$(shell $(GO) env GOARCH).tar.gz

help: ## СПРАВКА | Все публичные команды
	@awk 'BEGIN { FS=":.*## " } /^[a-zA-Z0-9_-]+:.*## / { split($$2, p, "[|]"); g=p[1]; sub(/ +$$/, "", g); sub(/^ +/, "", p[2]); if (!(g in known)) { order[++n]=g; known[g]=1 } text[g]=text[g] sprintf("  make %-25s %s\n", $$1, p[2]) } END { for (i=1;i<=n;i++) printf "\n%s\n%s", order[i], text[order[i]] }' $(MAKEFILE_LIST)
	@printf '\nBENCHTIME=100ms BENCHCOUNT=3 BENCHCPU=1 BENCHVIEW=compact\n'
	@printf 'BENCHSIZE=0 — все n/chunks/cap; BENCHOP= — все операции; BENCHDIR=bin/benchmarks.\n'
	@printf 'Пример: make bench-list BENCHSIZE=1000 BENCHOP=search-missing BENCHTIME=20ms\n'
	@printf 'Все таблицы: BENCHVIEW=full. Тщательнее: BENCHTIME=500ms BENCHCOUNT=5.\n'
	@printf 'Raw: make bench-list-raw. Проверка веток: make bench-smoke (1x, не скорость).\n'

deps-check: ## ЗАВИСИМОСТИ | Скачать и проверить зависимости
	$(GO) mod download
	$(GO) mod verify
mod-check: ## ЗАВИСИМОСТИ | go mod tidy и проверка изменений go.mod/go.sum
	$(GO) mod tidy
	@if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then git diff --exit-code -- go.mod go.sum; fi

test: test-unit test-integration ## ТЕСТЫ | Unit + integration

test-integration: ## ТЕСТЫ | Запуск семи исходных cmd и проверка JSON
	$(GO) test ./test/integration/...

run-all: ## ЗАПУСК | Все семь исходных примеров
	@set -e; for cmd in $(CMDS); do echo "== $$cmd =="; $(GO) run ./cmd/$$cmd; done
run-01_linkedlist: ## ЗАПУСК | Singly/Doubly, reverse и связи
run-02_stack: ## ЗАПУСК | LIFO на slice/list
run-03_queue: ## ЗАПУСК | FIFO на slice/list
run-04_buffer: ## ЗАПУСК | Linear Buffer и bytes.Buffer
run-05_ringbuffer: ## ЗАПУСК | Wrap-around, full/empty
run-06_hashtable: ## ЗАПУСК | Коллизии и rehash
run-07_visualizations: ## ЗАПУСК | JSON шести исходных сцен
$(addprefix run-,$(CMDS)):
	$(GO) run ./cmd/$(patsubst run-%,%,$@)

build: ## СБОРКА | Семь примеров и benchreport в bin
	@mkdir -p bin
	@set -e; for cmd in $(BUILD_CMDS); do $(GO) build -buildvcs=false -o bin/$$cmd ./cmd/$$cmd; done
package: build ## СБОРКА | Архив текущих исполняемых файлов
	tar -czf $(PACKAGE_FILE) -C bin $(BUILD_CMDS)
coverage-check: coverage ## КАЧЕСТВО | Проверить порог покрытия 90%
	@coverage="$$($(GO) tool cover -func=$(COVERAGE_FILE) | awk '/^total:/ {gsub("%", "", $$3); print $$3}')"; awk -v c="$$coverage" -v t="$(COVERAGE_THRESHOLD)" 'BEGIN {printf "Coverage %.1f%% (threshold %.1f%%)\n",c,t;exit(c<t)}'
ci: deps-check mod-check compile fmt-check vet test-unit test-integration test-race coverage-check build ## КАЧЕСТВО | Полная локальная проверка без длинных измерений

bench-list: ## БЕНЧМАРКИ | List/Slice: поиск, обход, build, изменения
bench-stack: ## БЕНЧМАРКИ | Вся серия n Push + n Pop
bench-queue: ## БЕНЧМАРКИ | 64 пары FIFO при постоянной длине n
bench-hash: ## БЕНЧМАРКИ | Build, lookup, тройка и восстановление
bench-buffer: ## БЕНЧМАРКИ | Накопление отдельно от bounded ring
bench-memory: ## БЕНЧМАРКИ | Allocation build, не retained heap
bench-list-raw: ## СЫРОЙ ВЫВОД | Исходный testing.B: list
bench-stack-raw: ## СЫРОЙ ВЫВОД | Исходный testing.B: stack
bench-queue-raw: ## СЫРОЙ ВЫВОД | Исходный testing.B: queue
bench-hash-raw: ## СЫРОЙ ВЫВОД | Исходный testing.B: hash
bench-buffer-raw: ## СЫРОЙ ВЫВОД | Исходный testing.B: buffer
bench-memory-raw: ## СЫРОЙ ВЫВОД | Исходный testing.B: memory
$(BENCH_TARGETS) $(addsuffix -raw,$(BENCH_TARGETS)):
	@$(GO) run ./cmd/benchreport -go "$(GO)" -kind $(patsubst bench-%,%,$(patsubst %-raw,%,$@)) -benchtime "$(BENCHTIME)" -count $(BENCHCOUNT) -cpu "$(BENCHCPU)" -view "$(BENCHVIEW)" -size $(BENCHSIZE) -op "$(BENCHOP)" -dir "$(BENCHDIR)" $(if $(filter %-raw,$@),-raw,)

# Даже make -j выполняет эти группы последовательно внутри одной recipe.
bench-all: ## БЕНЧМАРКИ | Все шесть групп последовательно
	@set -e; for group in $(BENCH_GROUPS); do $(MAKE) --no-print-directory bench-$$group; done
bench-all-raw: ## СЫРОЙ ВЫВОД | Все шесть групп последовательно
	@set -e; for group in $(BENCH_GROUPS); do $(MAKE) --no-print-directory bench-$$group-raw; done
bench: bench-all ## БЕНЧМАРКИ | Псевдоним bench-all
bench-smoke: ## ТЕСТЫ | Один вызов всех benchmark-сценариев, без оценки скорости
	$(MAKE) --no-print-directory bench-all BENCHTIME=1x BENCHCOUNT=1 BENCHCPU=1

.PHONY: fmt fmt-check vet compile test-unit test-race coverage clean help deps-check mod-check test test-integration run-all run-01_linkedlist run-02_stack run-03_queue run-04_buffer run-05_ringbuffer run-06_hashtable run-07_visualizations build package coverage-check ci bench-list bench-stack bench-queue bench-hash bench-buffer bench-memory bench-list-raw bench-stack-raw bench-queue-raw bench-hash-raw bench-buffer-raw bench-memory-raw bench-all bench-all-raw bench bench-smoke
