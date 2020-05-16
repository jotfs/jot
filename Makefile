protos:
	cp ../iotafs-server/internal/protos/upload/*.proto internal/protos/upload
	protoc internal/protos/upload/upload.proto --twirp_out=. --go_out=.

