#!/bin/bash
cd $(cd $(dirname $0); pwd)

./stop.sh
./run.sh
