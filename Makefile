release:
	@set GOOS=windows
	@set GOARCH=amd64
	@go build -o win.exe cmd/main.go
	@set GOOS=darwin
	@go build -o mac cmd/main.go