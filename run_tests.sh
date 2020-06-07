rm -f jotfs-test.db
docker run --rm --name jot-testing -p 9003:9000 -d minio/minio server /tmp/data
AWS_ACCESS_KEY_ID=minioadmin AWS_SECRET_ACCESS_KEY=minioadmin aws --endpoint-url http://localhost:9003 s3 mb s3://jotfs-test
./bin/jotfs -config jotfs.test.toml -debug &
jot_pid=$!
go test -race -coverprofile=coverage.txt -covermode=atomic ./...
kill $jot_pid
docker stop jot-testing
rm -f jotfs-test.db