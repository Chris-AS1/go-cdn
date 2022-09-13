package main

import (
	"fmt"
	"go-cdn/utils"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

var (
	dataFolder = "./resources"
)

// Root Handle - version number
func RootHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "API v1")
}

func getFiles(dir string) map[int]string {
	files, err := os.ReadDir(dataFolder)
	var ret = make(map[int]string)

	if err != nil {
		log.Fatal(err)
	}

	for i, file := range files {
		ret[i] = file.Name()
	}

	return ret
}

// Testing - Lists files on a directory
func ListHandler(w http.ResponseWriter, r *http.Request) {
	// w.WriteHeader(http.StatusOK)
	// for k, v := range getFiles(dataFolder) {
	// 	io.WriteString(w, strconv.Itoa(k)+" "+v+"\n")

	// }

	w.WriteHeader(http.StatusOK)
	for k, v := range GetImageList() {
		io.WriteString(w, strconv.Itoa(k)+" "+v+"\n")
	}

}

// Returns a specified image
func ImageHandler(w http.ResponseWriter, r *http.Request) {
	log.Print(r.URL)
	vars := mux.Vars(r)
	img_id := vars["id"]

	// If empty ID
	if img_id == "null" || img_id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	current_file_map := getFiles(dataFolder)
	img_id_int, err := strconv.Atoi(img_id)

	// If atoi fails
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Print(err)
		return
	}

	_, ok := current_file_map[img_id_int]

	// If NOT in map
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Can't access ID: %s [%d]", img_id, img_id_int)
		return
	}

	buff, err := os.ReadFile(fmt.Sprintf("%s/%s", dataFolder, current_file_map[img_id_int]))

	// If read error
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Print(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "image/jpg")
	w.Write(buff)
}

func main() {
	utils.LoadEnv()

	log.Print("Starting Server")

	r := mux.NewRouter().StrictSlash(true)
	r.HandleFunc("/", RootHandler)

	// Serving Path Selection
	b, err := strconv.ParseBool(utils.EnvSettings.DeliveringSubPathEnable)
	if err != nil {
		log.Panic(err)
	}

	if b {
		log.Printf("Serving Path: /%s/{id}/", utils.EnvSettings.DeliveringSubPath)

		r.HandleFunc(fmt.Sprintf("/%s", utils.EnvSettings.DeliveringSubPath), ImageHandler)
		r.HandleFunc(fmt.Sprintf("/%s/{id}", utils.EnvSettings.DeliveringSubPath), ImageHandler)
	} else {
		log.Print("Serving Path: /{id}/")
		r.HandleFunc("/", ImageHandler)
		r.HandleFunc("/{id:[0-9]+}", ImageHandler)
	}

	r.HandleFunc("/list/", ListHandler)
	http.Handle("/", r)

	srv := &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf(":%s", utils.EnvSettings.DeliveringPort),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("Serving Port: %s", utils.EnvSettings.DeliveringPort)
	log.Fatal(srv.ListenAndServe())
}
