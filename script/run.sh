#!/bin/bash

cd $(cd $(dirname $0); pwd)

while [ "$(ps -e > /tmp/psi && grep "proto_trace" /tmp/psi)" != "" ]
do
    sleep 1
done

cd ..
env GOTRACEBACK=crash nohup ./proto_trace > proto_trace.log 2>&1 &