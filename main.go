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
	fileMap              map[string]string
	ticker               = time.NewTicker(5 * time.Second)
	quit                 = make(chan struct{})
)

// Root Handle - Version Number
func RootHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "API v1")
}

// Lists files on a directory
func GetListHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Authenticate

	w.WriteHeader(http.StatusOK)
	for k, v := range fileMap {
		io.WriteString(w, k+" "+v+"\n")
	}
}

// Builds the correct path given the filename
func getImagePath(filename string) string {
	return fmt.Sprintf("%s/%s", dataFolder, filename)
}

func readImage(path string) (bool, []byte) {
	buff, err := os.ReadFile(path)

	// If read error
	if err != nil {
		log.Print(err)
		return false, nil
	}

	return true, buff
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

	// Gets filename from ID
	_, ok := fileMap[img_id]

	// If ID NOT in map
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ResponseInvalidID)
		log.Printf("Can't access ID: %s", img_id)
		return
	}

	// Cheks if Redis is enabled on env
	b, err := strconv.ParseBool(utils.EnvSettings.RedisEnable)
	if err != nil {
		log.Panic(err)
	}

	var outBuf []byte
	cache_ok := false

	if b {
		log.Printf("[%s] %s: Hit Times [%d]", r.Method, r.URL, recordAccess(img_id))

		// Checks if ID is in cache
		// TODO Wrap into single function because of double read from disk at first hit
		cache_ok, outBuf = getFromCache(img_id)
	}

	if !cache_ok {
		var read_ok bool
		read_ok, outBuf = readImage(getImagePath(fileMap[img_id]))

		if !read_ok {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ResponseInvalidImage)
			return
		}
	} else {
		log.Printf("[%s] %s: Got From Cache [%s]", r.Method, r.URL, img_id)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "image/jpg")
	w.Write(outBuf)
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

	// Gets filename from ID
	_, ok := fileMap[img_id]

	// If ID NOT in map
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Can't access ID: %s", img_id)
		return
	}

	err := os.Remove(getImagePath(fileMap[img_id]))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{false, "error deleting file"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ResponseSuccess)
}

func refreshClock() {
	for {
		select {
		case <-ticker.C:
			refreshCache()
		case <-quit:
			ticker.Stop()
			return
		}
	}
}

func main() {
	utils.LoadEnv()
	fileMap = BuildFileMap()

	log.Printf("Redis connection: %s", ConnectRedis())
	log.Print("Starting Server")

	// go refreshClock()

	r := mux.NewRouter().StrictSlash(true)

	// Disabled
	// r.HandleFunc("/", RootHandler)

	// Serving Image Path
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

	// Serve List Path
	r.HandleFunc("/list/", GetListHandler)

	// Use Router
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
