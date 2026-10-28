.PHONY: all flash build size version install_tinygo_edit test-format test-displays test-unit check-images render-architecture
# The device command and internal/platform require TinyGo.
DEVICE ?= pico_co2
TARGET ?= pico

LINT_PKGS = ./internal/... ./pkg/... ./cmd/virtualdisplay

vi:
	tinygo-edit --target pico --editor nvim --wait

flash:
	tinygo flash -size=short -target=$(TARGET) -monitor ./cmd/$(DEVICE)/

build:
	tinygo build -size=full -target=$(TARGET) -o main.elf ./cmd/$(DEVICE)/

size:
	go tool nm -size main.elf | sort -k 2,2 -nr | head -n 20

version:
	go version
	tinygo version
	tinygo-edit --version

install_tinygo_edit:
	go install github.com/sago35/tinygo-edit@latest

displays:
	go run ./cmd/virtualdisplay/
	git status --short images/

test-format:
	test -z "$$(gofmt -l cmd internal pkg)"

test-unit:
	go test ./internal/... ./pkg/... ./cmd/virtualdisplay

check-images: test-displays
	test -z "$$(git status --porcelain -- images/)"

lint:
	@echo "Running golangci-lint..."
	golangci-lint run --config .golangci.yaml $(LINT_PKGS)

lint-fix:
	@echo "Running golangci-lint autofix..."
	golangci-lint run --fix --config .golangci.yaml $(LINT_PKGS)

lint-fmt:
	golangci-lint fmt --config .golangci.yaml
