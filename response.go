package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

func WriteResponse(writer http.ResponseWriter, format string, a ...interface{}) {
	_, err := fmt.Fprintf(writer, format, a)
	if err != nil {
		log.Print("WriteResponse Error", err.Error())
	}
}

func WriteByteResponse(writer http.ResponseWriter, format []byte, a ...interface{}) {
	WriteResponse(writer, string(format), a)
}

func Template(writer http.ResponseWriter, file string, data interface{}) {
	page, _ := template.ParseFiles(file)
	err := page.Execute(writer, data)
	if err != nil {
		log.Print("WriteResponse Error", err.Error())
	}
}

func JSONResult(msg string) string {
	return fmt.Sprintf("{\"result\": \"%s\"}", msg)
}
