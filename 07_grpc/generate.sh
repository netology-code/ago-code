#!/bin/bash

mkdir -p pkg/event/v1

protoc -Iapi/proto/v1 -Ithird_party --go_out=plugins=grpc:pkg/event/v1 --go_opt=paths=source_relative api/proto/v1/event.proto