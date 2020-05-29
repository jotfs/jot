protos:
	cp ../iotafs-server/internal/protos/*.proto internal/protos
	protoc internal/protos/api.proto --twirp_out=. --go_out=.
