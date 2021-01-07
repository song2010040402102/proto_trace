#!/bin/bash

cd $(cd $(dirname $0); pwd)

fmt_dir(){
    for file in `ls $1`
    do
        if [ -d $1"/"$file ]; then
            if [ $file != Lyingdragon_Protocol ]; then
                fmt_dir $1"/"$file
            fi
        else
            if [[ $file == *.go ]]; then
                go fmt $1"/"$file
                echo $1"/"$file
            fi
        fi
    done
}
fmt_dir ..