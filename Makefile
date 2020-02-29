.PHONY: mac
win:
	@echo "build windows at build/win_64.exe"
	@GOOS=windows GOARCH=amd64 go build -o build/flowers_win_amd64.exe flowers/main.go

.PHONY: mac
mac:
	@echo "build windows at build/win_64.exe"
	@GOOS=windows GOARCH=amd64 go build -o build/flowers_osx_amd64 flowers/main.go
.PHONY: linux
linux:
	@echo "build windows at build/flowers_linux_amd64"
	@GOOS=linux GOARCH=amd64 go build -o build/flowers_linux_amd64 flowers/main.go

all: win mac linux
