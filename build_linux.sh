mkdir -p build/bin

env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/bin/s3download jobs/s3download/main.go
cp conf/*_ec2.yaml build/bin/

CROSS_HOST=114.67.106.133
CROSS_PATH=/root/projects
scp -r build/* root@$CROSS_HOST:$CROSS_PATH