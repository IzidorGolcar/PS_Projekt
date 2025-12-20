go generate ./...
echo building data service

mkdir build

go build -o build/data_service cmd/data/main.go
