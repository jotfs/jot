protos:
	# Pull the main jotfs repo to get the protobuf definitions and compile them
	mkdir -p build
	rm -rf ./build/jotfs
	git clone --depth=1 https://github.com/jotfs/jotfs ./build/jotfs
	cp ./build/jotfs/internal/protos/*.proto internal/protos
	protoc internal/protos/api.proto --twirp_out=. --go_out=.

dl-server-binary:
	# Download the latest JotFS binary release from github and unzip it to ./bin/jotfs
	bash download_jotfs_binary.sh

protos-dev:
	# Build the protobufs from local files. For development only.
	cp ../jotfs-server/internal/protos/*.proto internal/protos
	protoc internal/protos/api.proto --twirp_out=. --go_out=.

tests:
	bash run_tests.sh