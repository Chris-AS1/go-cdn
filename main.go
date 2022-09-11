package main

import (
	"errors"
	"fmt"
	"io"
	"time"
	"io/ioutil"
	"net/http"
	"log"
	"github.com/gorilla/mux"
)

var (
	ResourceNotFoundException = errors.New("resource not found")
)

// Root Handle - version number
func RootHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "API v1")
}

// Testing - Lists files on a directory
func ListHandler(w http.ResponseWriter, r *http.Request) {
	files, err := ioutil.ReadDir("resources/")
	vars := mux.Vars(r)

	log.Print(vars)
    if err != nil {
		log.Fatal(err)
    }
	
	var str string

	for _, file := range files {
		str += file.Name() + "\n"
	}
	
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, str)
	return
}

// Returns a specified image
func ImageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)	
	img_id := vars["id"]

	if img_id != "null" && img_id != "" {
		buff, err := ioutil.ReadFile(fmt.Sprintf("resources/%s.jpg", vars["id"]))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Print(err)
			return
		}
	
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "image/jpg")
		w.Write(buff)
		return
	}

	w.WriteHeader(http.StatusBadRequest)
	w.Header().Set("Content-Type", "application/text")
	io.WriteString(w, "EMPTY!\n")
	return
}

func main() {
	log.Print("Starting Server")

	r := mux.NewRouter().StrictSlash(true)
	r.HandleFunc("/", RootHandler)

	r.HandleFunc("/image/", ImageHandler)
	r.HandleFunc("/image/{id}", ImageHandler)

	r.HandleFunc("/list/", ListHandler)

	http.Handle("/", r)
	
	srv := &http.Server{
        Handler:      r,
        Addr:         "127.0.0.1:3333",
        WriteTimeout: 15 * time.Second,
        ReadTimeout:  15 * time.Second,
    }

    log.Fatal(srv.ListenAndServe())
}