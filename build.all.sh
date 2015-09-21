echo "building for darwin 386..."
env GOOS=darwin GOARCH=386 go build -o ch-darwin-386 main.go 
echo "done."
echo "building for darwin amd64..."
env GOOS=darwin GOARCH=amd64 go build -o ch-darwin-amd64 main.go 
echo "done."
echo "building for windows amd64..."
env GOOS=windows GOARCH=amd64 go build -o ch-windows-amd64 main.go 
echo "done."
echo "building for linux amd64..."
env GOOS=linux GOARCH=amd64 go build -o ch-linux-amd64 main.go
echo "All done."

