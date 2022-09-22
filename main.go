package main

import (
	"encoding/json"
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

type GenericImage struct {
	Data string `json:"data"`
}

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

var (
	dataFolder           = "./resources"
	ResponseSuccess      = Response{Success: true}
	ResponseInvalidImage = Response{Success: false, Message: "invalid image"}
	ResponseInvalidID    = Response{Success: false, Message: "invalid ID"}
)

// Root Handle - Version Number
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
func GetListHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	for k, v := range getFiles(dataFolder) {
		io.WriteString(w, strconv.Itoa(k)+" "+v+"\n")

	}

	// w.WriteHeader(http.StatusOK)
	// for k, v := range GetImageList() {
	// 	io.WriteString(w, k+" "+v+"\n")
	// }

}

// Builds the correct path given the filename
func getImagePath(filename string) string {
	return fmt.Sprintf("%s/%s", dataFolder, filename)
}

// Returns a specified image
func GetImageHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[%s] %s", r.Method, r.URL)
	vars := mux.Vars(r)
	img_id := vars["id"]

	// If empty ID
	if img_id == "null" || img_id == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ResponseInvalidID)
		return
	}

	current_file_map := getFiles(dataFolder)
	img_id_int, err := strconv.Atoi(img_id)

	// If atoi fails (invalid ID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Print(err)
		return
	}

	_, ok := current_file_map[img_id_int]

	// If ID NOT in map
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ResponseInvalidID)
		log.Printf("Can't access ID: %s [%d]", img_id, img_id_int)
		return
	}

	buff, err := os.ReadFile(getImagePath(current_file_map[img_id_int]))

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

// Adds image - TODO base64 + write file
func PostImageHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[%s] %s", r.Method, r.URL)

	var img_to_add GenericImage
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&img_to_add)
	if err != nil {
		log.Panic(err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ResponseInvalidImage)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ResponseSuccess)
}

// Deletes an image from disk
func DeleteImageHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[%s] %s", r.Method, r.URL)
	vars := mux.Vars(r)
	img_id := vars["id"]

	// If empty ID
	if img_id == "null" || img_id == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ResponseInvalidID)
		return
	}

	current_file_map := getFiles(dataFolder)
	img_id_int, err := strconv.Atoi(img_id)

	// If atoi fails (invalid ID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Print(err)
		return
	}

	_, ok := current_file_map[img_id_int]

	// If ID NOT in map
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Can't access ID: %s [%d]", img_id, img_id_int)
		return
	}

	err = os.Remove(getImagePath(current_file_map[img_id_int]))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{false, "error deleting file"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ResponseSuccess)
}

func main() {
	utils.LoadEnv()

	log.Print("Starting Server")

	r := mux.NewRouter().StrictSlash(true)

	// Disabled
	// r.HandleFunc("/", RootHandler)

	// Serving Path Selection
	b, err := strconv.ParseBool(utils.EnvSettings.DeliveringSubPathEnable)
	if err != nil {
		log.Panic(err)
	}

	if b {
		log.Printf("Serving Path: /%s/{id}/", utils.EnvSettings.DeliveringSubPath)

		url, url_id := fmt.Sprintf("/%s", utils.EnvSettings.DeliveringSubPath),
			fmt.Sprintf("/%s/{id}", utils.EnvSettings.DeliveringSubPath)

		r.HandleFunc(url, GetImageHandler).Methods("GET")
		r.HandleFunc(url_id, GetImageHandler).Methods("GET")

		// Check if insertion endpoint is enabled
		add, err := strconv.ParseBool(utils.EnvSettings.EnableInsertion)
		if add {
			r.HandleFunc(url, PostImageHandler).Methods("POST")
		}

		if err != nil {
			log.Panic(err)
		}

		// Check if deletion endpoint is enabled
		del, err := strconv.ParseBool(utils.EnvSettings.EnableDeletion)

		if del {
			r.HandleFunc(url_id, DeleteImageHandler).Methods("DELETE")
		}

		if err != nil {
			log.Panic(err)
		}
	} else {
		log.Print("Serving Path: /{id}/")
		r.HandleFunc("/", GetImageHandler).Methods("GET")
		r.HandleFunc("/{id:[0-9]+}", GetImageHandler).Methods("GET")
	}

	r.HandleFunc("/list/", GetListHandler)
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
