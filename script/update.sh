#!/bin/bash
set -e

cd $(cd $(dirname $0); pwd)

echo "download package..."
wget http://10.0.253.27:9111/proto_trace.tgz -O ../../proto_trace.tgz

echo "stop server..."
./stop.sh

echo "unzip package..."
cd ..
mv config.yaml ../
cd ..
rm -rf proto_trace
tar -zxf proto_trace.tgz
rm -f proto_trace.tgz
mv config.yaml proto_trace

echo "run server..."
cd proto_trace/script
./run.sh

echo "done!"
