.PHONY: dev build clean test frontend-install

dev:
	wails3 dev

build:
	cd frontend && npm run build
	wails3 build

build-windows:
	cd frontend && npm run build
	wails3 build -platform windows/amd64

build-darwin:
	cd frontend && npm run build
	wails3 build -platform darwin/universal

clean:
	rm -rf frontend/dist frontend/node_modules build

test:
	go test ./...

frontend-install:
	cd frontend && npm install
