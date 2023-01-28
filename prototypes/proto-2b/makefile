all:

deps:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

protoc:
	mkdir -p pbtypes
	protoc -I=proto proto/headers.proto  --go_out=./pbtypes --go_opt=paths=source_relative
	protoc -I=proto proto/metadata.proto  --go_out=./pbtypes --go_opt=paths=source_relative

clean:
	rm -rf pbtypes
