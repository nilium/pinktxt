#!/usr/bin/env bash

here="$(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$here"

if ! [[ -d google-protobuf ]] ; then
	git clone --depth=1 --branch=v2.6.1 https://github.com/google/protobuf.git google-protobuf
fi

protoc -I"$here/google-protobuf/src" -I"$here" --gogo_out="marshaler=true,unmarshaler=true:." "$here/google-protobuf/src/google/protobuf/"descriptor.proto
protoc -I"$here/google-protobuf/src" -I"$here" --gogo_out="marshaler=true,unmarshaler=true:." "$here/google-protobuf/src/google/protobuf/"compiler/plugin.proto
# sed -i '' -e 's/^import google_protobuf /\/\/&/' google/protobuf/compiler/plugin.pb.go
gofmt -w -r '"google/protobuf" -> "github.com/nilium/pinktxt/internal/plugin/google/protobuf"' google/protobuf/compiler/plugin.pb.go

