package main

import (
	"context"
	"embed"
	_ "embed"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/psanford/memfs"
)

//go:embed images
var imagefs embed.FS

//go:embed static
var staticfs embed.FS

var mfs = memfs.New()

func main() {
	err := createThumbnails(imagefs, mfs)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	staticFiles, err := fs.Sub(staticfs, "static")
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	router := mux.NewRouter()
	router.Use(handlers.ProxyHeaders)
	router.Use(handlers.CompressHandler)
	router.Use(NewLoggingHandler(os.Stdout))
	router.PathPrefix("/thumbnails/").Handler(thumbnailHandler(http.StripPrefix("/thumbnails/", http.FileServer(http.FS(mfs)))))
	router.PathPrefix("/").Handler(http.FileServer(http.FS(staticFiles)))

	log.Print("Starting server.")
	server := http.Server{
		Addr:    ":8080",
		Handler: router,
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

func NewLoggingHandler(dst io.Writer) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return handlers.LoggingHandler(dst, h)
	}
}

func thumbnailHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if strings.Contains(request.Header.Get("Accept"), "image/webp") {
			webp := request.URL.Path + ".webp"
			_, err := mfs.Open(strings.TrimPrefix(webp, "/thumbnails/"))
			if err == nil {
				request.URL.Path = webp
			}
		}
		next.ServeHTTP(writer, request)
	})
}