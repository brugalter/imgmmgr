package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/a-h/templ"
	"github.com/brugalter/imgmmgr/view"
	"github.com/google/uuid"
)

func Upload(w http.ResponseWriter, r *http.Request) (err error) {
	uuid, err := uuid.NewV7()
	if err != nil {
		return err
	}

	dir := filepath.Join(filepath.Join("./files"), uuid.String())
	err = os.Mkdir(dir, 0750)
	if err != nil {
		return err
	}

	err = view.Code(dir, uuid.String()).Render(context.Background(), w)

	err = r.ParseMultipartForm(16384)
	if err != nil {
		return err
	}

	for _, files := range r.MultipartForm.File {
		for _, file := range files {

			path := filepath.Join(dir, file.Filename)
			dst, err := os.Create(path)
			if err != nil {
				return err
			}
			defer dst.Close()

			f, err := file.Open()
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(dst, f)
			switch {
			case strings.HasSuffix(path, "png"):
				err = view.File(path).Render(context.Background(), w)
			case strings.HasSuffix(path, "mp4"):
				err = view.Video(path, "video/mp4").Render(context.Background(), w)
			}

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				io.WriteString(w, err.Error())
			}
		}
	}

	return nil

}

func HandleUpload(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		err := Upload(w, r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, err.Error())
		}

		//w.WriteHeader(http.StatusCreated)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, "Method not allowed")
	}
}

func main() {
	base := view.Base("imgmmgr")
	button := view.Button()

	port := flag.String("port", "9000", "listen port")
	ip := flag.String("ip", "127.0.0.1", "Bind ip")

	flag.Parse()

	mux := http.NewServeMux()
	mux.Handle("/", templ.Handler(base))
	mux.Handle("/button", templ.Handler(button))
	mux.HandleFunc("/upload", HandleUpload)
	mux.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir("./files/"))))

	fmt.Println("listening on " + *ip + ":" + *port)
	err := http.ListenAndServe(*ip+":"+*port, mux)
	if errors.Is(err, http.ErrServerClosed) {
		log.Fatal("Server closed")
	} else if err != nil {
		log.Fatal("", err)
		os.Exit(1)
	}
}
