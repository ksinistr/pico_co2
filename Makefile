.PHONY: all flash build size version install_tinygo_edit flash_ens160_example build_ens160_example test-format test-displays test-unit check-images
# cmd/pico_co2 imports "machine" (TinyGo-only)
LINT_PKGS = ./internal/... ./pkg/... ./cmd/virtualdisplay

vi:
	tinygo-edit --target pico --editor nvim --wait

flash:
	tinygo flash -size=short -target=pico -monitor ./cmd/pico_co2/

build:
	tinygo build -size=full -target=pico -o main.elf ./cmd/pico_co2/

size:
	go tool nm -size main.elf | sort -k 2,2 -nr | head -n 20

version:
	go version
	tinygo version
	tinygo-edit --version

install_tinygo_edit:
	go install github.com/sago35/tinygo-edit@latest

flash_ens160_example:
	tinygo flash -target=pico -monitor ./pkg/ens160/example/

build_ens160_example:
	tinygo build -target=pico -o main.uf2 ./pkg/ens160/example/

test-displays:
	go run ./cmd/virtualdisplay/
	git status --short images/

test-format:
	test -z "$$(gofmt -l cmd internal pkg)"

test-unit:
	go test -v ./internal/app ./internal/clockedit ./internal/display ./internal/rtc ./internal/types ./internal/types/status ./pkg/fifo ./pkg/font ./pkg/layout ./pkg/widget ./pkg/miniplot ./pkg/sparkline

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
