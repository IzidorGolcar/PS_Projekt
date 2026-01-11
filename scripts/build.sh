echo 'building project'

go install golang.org/x/tools/cmd/stringer@latest
go mod tidy
go generate ./...

mkdir -p build

echo 'building data service'
go build -o build/data_service cmd/data/main.go

echo 'building TUI client'
go build -o build/client_tui cmd/tui_client/main.go

echo 'building CLI client'
go build -o build/client_cli cmd/cli_client/main.go

echo 'building control service'
go build -o build/control_service cmd/control/main.go


echo 'done'
