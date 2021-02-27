package main

import (
	"context"
	"embed"
	_ "embed"
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/acoshift/middleware"
	"github.com/blang/vfs/memfs"
)

//go:embed images
var imagefs embed.FS

//go:embed static
var staticfs embed.FS

var mfs = memfs.Create()

func main() {
	err := createThumbnails(imagefs, mfs)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/css", cssHandler)
	mux.HandleFunc("/favicon", faviconHandler)
	mux.HandleFunc("/thumbnails/", thumbnailHandler)

	h := middleware.Chain(
		requestLogger(),
		middleware.Compress(middleware.DeflateCompressor),
		middleware.Compress(middleware.GzipCompressor),
		middleware.Compress(middleware.BrCompressor),
	)(mux)

	log.Print("Starting server.")
	server := http.Server{
		Addr:    ":8080",
		Handler: h,
	}
	go func() {
		_ = server.ListenAndServe()
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Unable to shutdown: %s", err.Error())
	}
	log.Print("Finishing server.")
}

func indexHandler(writer http.ResponseWriter, request *http.Request) {
	if request.URL.Path == "/" {
		fileServerHandler(writer, request, "static/index.html", "text/html; charset=utf-8")
	} else {
		writer.WriteHeader(http.StatusNotFound)
	}
}

func cssHandler(writer http.ResponseWriter, request *http.Request) {
	fileServerHandler(writer, request, "static/main.css", "text/css; charset=utf-8")
}

func faviconHandler(writer http.ResponseWriter, request *http.Request) {
	fileServerHandler(writer, request, "static/favicon.ico", "image/x-icon; charset=utf-8")
}

func fileServerHandler(writer http.ResponseWriter, _ *http.Request, filename string, contentType string) {
	data, err := staticfs.ReadFile(filename)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", contentType)
	_, _ = writer.Write(data)
}

func thumbnailHandler(writer http.ResponseWriter, request *http.Request) {
	if strings.Contains(request.Header.Get("Accept"), "image/webp") {
		webp := request.URL.Path + ".webp"
		_, err := mfs.Stat(fmt.Sprintf("./%s", webp))
		if err == nil {
			request.URL.Path = webp
		}
	}
	request.URL.Path = strings.TrimPrefix(request.URL.Path, "/thumbnails/")
	f, err := mfs.OpenFile(request.URL.Path, os.O_RDWR, 0666)
	if err != nil {
		writer.WriteHeader(http.StatusNotFound)
		_, _ = writer.Write([]byte("thumbnail not found"))
		return
	}
	bytes, err := io.ReadAll(f)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		_, _ = writer.Write([]byte("error reading thumbnail"))
		return
	}
	_, _ = writer.Write(bytes)

}

func requestLogger() middleware.Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requesterIP := r.RemoteAddr
			log.Printf(
				"%s\t\t%s\t\t%s\t",
				requesterIP,
				r.Method,
				r.RequestURI,
			)
			h.ServeHTTP(w, r)
		})
	}
}
