
.PHONY: all macos linux

all: macos linux install

macos:
	@echo "Compiling macos binaries"
	env GOOS=darwin GOARCH=amd64 go build -o dist/macos/amd64/migrant
	chmod a+x ./dist/macos/amd64/migrant

linux:
	@echo "Compiling linux binaries"
	env GOOS=linux GOARCH=amd64 go build -o dist/linux/amd64/migrant
	chmod a+x ./dist/macos/amd64/migrant

install:
	cp ./dist/macos/amd64/migrant /usr/local/bin/migrant