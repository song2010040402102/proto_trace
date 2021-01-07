#!/bin/bash
cd $(cd $(dirname $0); pwd)

cd ..
go build -o ./proto_trace