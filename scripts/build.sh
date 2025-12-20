echo 'building project'
go generate ./...
echo 'building data service'
mkdir -p build
go build -o build/data_service cmd/data/main.go
echo 'done'
