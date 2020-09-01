mkdir -p build/bin

go build -o build/bin/s3download jobs/s3download/main.go
cp conf/*.yaml build/bin/
