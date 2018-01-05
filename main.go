package main

import (
	"bytes"
	"image"
	"image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/upload", generateGIFHandler)
	http.Handle("/", http.FileServer(http.Dir("static")))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

const defaultMaxMemory = 32 << 20 // 32 MB (comes from net/http)

func generateGIFHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseMultipartForm(defaultMaxMemory); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var g gif.GIF
	for _, f := range r.MultipartForm.File["files"] {
		img, err := readImage(f)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		g.Image = append(g.Image, img.(*image.Paletted))
		g.Delay = append(g.Delay, 15)
	}
	buf := new(bytes.Buffer)
	if err := gif.EncodeAll(buf, &g); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	io.Copy(w, buf)
}

func readImage(fg *multipart.FileHeader) (image.Image, error) {
	f, err := fg.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	return img, err
}
