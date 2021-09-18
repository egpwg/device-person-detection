#!/bin/bash
env GOOS=linux GOARCH=arm64 go build -o "person-detection-device-service" -ldflags "-w -s"

if [ ! -d bin ];then
    mkdir bin
fi
mv -f person-detection-device-service bin