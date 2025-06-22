.DEFAULT_GOAL=build
BINARY_NAME=ptt


.PHONY: dev 
dev:
	@echo "launching air for live reloading..."
	air


.PHONY: deps
deps:
	@echo "fixing dependencies..."
	go mod tidy
	go mod vendor


.PHONY: proto
proto:
	@echo "compiling protocol buffers..."
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative shared/proto/module.proto shared/proto/store.proto


.PHONY: templ
templ:
	@echo "generating templ files..."
	templ generate -path internal/template


.PHONY: tailwind
tailwind:
	@echo "processing tailwindcss..."
	./node_modules/.bin/tailwindcss -i ./static/css/input.css -o ./static/css/output.css


.PHONY: build
build: proto tailwind templ
	@echo "building go binary..."
	go build -o build/${BINARY_NAME} main.go


.PHONY: test
test:
	@echo "starting tests..."
	go test ./...


.PHONY: clean
clean:
	@echo "cleaning..."
	rm -rf ./build
	rm -rf ./tmp
	rm -rf ./plugins
