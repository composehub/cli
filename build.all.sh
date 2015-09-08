env GOOS=darwin GOARCH=amd64 go build -o dcm-darwin-amd64 main.go 
env GOOS=windows GOARCH=amd64 go build -o dcm-windows-amd64 main.go 
env GOOS=linux GOARCH=amd64 go build -o dcm-linux-amd64 main.go
