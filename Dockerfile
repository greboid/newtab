FROM registry.greboid.com/mirror/golang:latest as builder
WORKDIR /app
COPY go.mod /app
COPY go.sum /app
COPY static /app/static/
COPY images /app/images/
COPY thumbnails.go /app
COPY main.go /app
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o main .

FROM scratch
WORKDIR /app
COPY --from=builder /app/main /newtab-site
EXPOSE 8080
CMD ["/newtab-site"]
