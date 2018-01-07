package main

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	http.HandleFunc("/upload/", handler)
	http.Handle("/media/", http.StripPrefix("/media/", http.FileServer(http.Dir("media"))))
	http.Handle("/", http.FileServer(http.Dir("static")))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

type server struct{}

const defaultMaxMemory = 32 << 20 // 32 MB (comes from net/http)

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseMultipartForm(defaultMaxMemory); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var g gif.GIF
	var fname string
	for _, f := range r.MultipartForm.File["images[]"] {
		if fname == "" {
			fname = f.Filename[:len(f.Filename)-len(filepath.Ext(f.Filename))] + ".gif"
		}

		img, err := readImage(f)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		g.Image = append(g.Image, img.(*image.Paletted))
		g.Delay = append(g.Delay, 20)
	}
	buf := new(bytes.Buffer)
	if err := gif.EncodeAll(buf, &g); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	url, err := upload(fname, buf)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, url)
}

func readImage(fh *multipart.FileHeader) (image.Image, error) {
	f, err := fh.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	return img, err
}

func upload(fname string, content io.Reader) (string, error) {
	dir := filepath.Join("media", randString())
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	fpath := filepath.Join(dir, fname)
	f, err := os.Create(fpath)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(f, content); err != nil {
		return "", err
	}
	return "/" + fpath, nil
}

func randString() string {
	p := make([]byte, 16)
	rand.Read(p)
	return fmt.Sprintf("%x", p)
}
