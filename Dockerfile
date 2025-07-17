FROM golang:1.24 as builder
WORKDIR /app
COPY go.mod /app
COPY go.sum /app
COPY static /app/static/
COPY main.go /app
RUN CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o main .

FROM ghcr.io/greboid/dockerbase/nonroot:1.20250716.0
WORKDIR /app
COPY --from=builder --chown=65532:65532 /app/main /newtab-site
EXPOSE 8080
CMD ["/newtab-site"]
