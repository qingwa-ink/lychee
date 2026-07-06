.PHONY: build run dev vet test fmt clean

# 生产构建（SQLite 需 CGO）
build:
	CGO_ENABLED=1 go build -o bin/lychee ./cmd/server

# 直接运行（go run，开发用）
run:
	go run ./cmd/server

# 运行已构建的二进制
dev: build
	./bin/lychee

vet:
	go vet ./...

test:
	go test ./...

fmt:
	gofmt -w .

clean:
	rm -rf bin/
