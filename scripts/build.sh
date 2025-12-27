echo 'building project'
go generate ./...

mkdir -p build

echo 'building data service'
go build -o build/data_service cmd/data/main.go

echo 'building client'
go build -o build/client cmd/client/main.go

echo 'building control service'
go build -o build/contrl_service cmd/control/main.go

echo 'done'
