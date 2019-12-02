#!/bin/bash
export CC=arm-linux-gnueabihf-gcc
env GOOS=linux GOARCH=arm GOARM=5 CGO_ENABLED=1 go build
mv $PWD/golit $PWD/build/golit

timestamp=$(date +%s)
zip -r golit_$timestamp.zip $PWD/build
echo Done!
