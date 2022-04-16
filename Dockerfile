FROM reg.g5d.dev/golang as builder
WORKDIR /app
COPY go.mod /app
COPY go.sum /app
COPY static /app/static/
COPY images /app/images/
COPY thumbnails.go /app
COPY main.go /app
RUN CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o main .

FROM reg.g5d.dev/base
WORKDIR /app
COPY --from=builder --chown=65532:65532 /app/main /newtab-site
EXPOSE 8080
CMD ["/newtab-site"]