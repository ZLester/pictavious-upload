package main

import (
	"crypto/rand"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

const MAX_IMAGE_MB = 3
const MAX_IMAGE_SIZE = 1024 * 1024 * MAX_IMAGE_MB
const FILE_NAME_LENGTH = 16
const IMAGE_DIR = "./images"

var validFileTypes = map[string]bool{
	"image/jpeg": true,
	"image/jpg":  true,
	"image/gif":  true,
	"image/png":  true,
}

func createFileName(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return fmt.Sprintf("%x", bytes)
}

func createImage(w http.ResponseWriter, r *http.Request) {

	r.Body = http.MaxBytesReader(w, r.Body, MAX_IMAGE_SIZE)
	err := r.ParseMultipartForm(MAX_IMAGE_SIZE)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Image too large. Max file size of " + strconv.Itoa(MAX_IMAGE_MB) + "mb"))
		return
	}

	file, _, err := r.FormFile("image")

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid image."))
		return
	}

	defer file.Close()
	fileBytes, err := ioutil.ReadAll(file)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Could not read file."))
		return
	}

	contentType := http.DetectContentType(fileBytes)

	if _, ok := validFileTypes[contentType]; ok != true {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid image type."))
		return
	}

	fileName := createFileName(FILE_NAME_LENGTH)

	fmt.Println(contentType)

	fileExtensions, err := mime.ExtensionsByType(contentType)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Could not read file extension."))
		return
	}
	fmt.Println(fileExtensions)

	path := filepath.Join(IMAGE_DIR, fileName+fileExtensions[0])
	log.Printf("Saving %s", path)

	createdFile, err := os.Create(path)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error while saving file."))
		return
	}

	if _, err := createdFile.Write(fileBytes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error while creating file."))
		return
	}

	if err := createdFile.Close(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error while closing file."))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fileName + fileExtensions[0]))
}

func createDir(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.Mkdir(path, 0755)
		if err != nil {
			panic(err)
		}
	}
}

func main() {
	createDir(IMAGE_DIR)

	router := mux.NewRouter()

	router.
		Path("/images").
		Methods("POST").
		HandlerFunc(createImage)

	log.Printf("Pictavious Upload started on %s", os.Getenv("PORT"))
	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), router))
}
