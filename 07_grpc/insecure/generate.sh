#!/bin/bash

mkdir -p pkg/fine/v1

protoc -Iapi/proto/v1 -Ithird_party --go_out=plugins=grpc:pkg/fine/v1 --go_opt=paths=source_relative api/proto/v1/fine.proto