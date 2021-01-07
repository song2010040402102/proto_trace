#!/bin/bash
set -e

cd $(cd $(dirname $0); pwd)

echo "pull protocol from svn..."
cd ../Lyingdragon_Protocol
svn update

echo "build protocol..."
./build.sh

echo "build server..."
cd ../script
./make.sh

echo "pack..."
mkdir proto_trace && cd proto_trace
mkdir Lyingdragon_Protocol
cp ../../Lyingdragon_Protocol/*.proto Lyingdragon_Protocol
mkdir script
cp ../../script/run.sh script
cp ../../script/stop.sh script
cp ../../script/update.sh script
cp ../../config.yaml ./
cp ../../README ./
cp -r ../../web ./
cp ../../proto_trace ./
cd ..
tar -zcf proto_trace.tgz proto_trace
mv proto_trace.tgz ../../http_server/
rm -rf proto_trace

echo "done!"
