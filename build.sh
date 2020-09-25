mkdir -p build/bin

go build -o build/bin/s3download jobs/s3download/cmd/main.go
go build -o build/bin/scp jobs/scp/cmd/main.go
go build -o build/bin/dump-tables jobs/dump-tables/cmd/main.go

cp conf/*.yaml build/bin/