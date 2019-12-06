#!/bin/bash
export CC=arm-linux-gnueabihf-gcc
env GOOS=linux GOARCH=arm GOARM=5 CGO_ENABLED=1 go build
mv $PWD/golit $PWD/build/golit

echo Done!
