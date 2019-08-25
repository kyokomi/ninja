build:
	go build -o ninja
build_win:
	GOOS=windows GOARCH=amd64 go build -o ninja.exe

