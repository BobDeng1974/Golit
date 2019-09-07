package main

import (
	"io/ioutil"
)

type StaticFile struct {
	Data []byte
}

func load_static(path string) (*StaticFile, error) {
	buffer, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return &StaticFile{Data: buffer}, nil
}
