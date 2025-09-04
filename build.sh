GOOS=android GOARCH=arm64 go build -o build/ablage-android-arm64 .
GOOS=linux GOARCH=amd64 go build -o build/ablage-linux-amd64 .
GOOS=windows GOARCH=amd64 go build -o build/ablage-windows-amd64.exe .
