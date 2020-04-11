FROM golang:latest AS development
RUN git clone --progress --verbose --depth=1 https://github.com/Bpazy/rss-middleware /rss-middleware
WORKDIR /rss-middleware
RUN go env && CGO_ENABLED=0 go build ./cmd/rss-torrent

FROM alpine:latest AS production
COPY --from=development /rss-middleware/rss-torrent /rss-middleware/rss-torrent
WORKDIR /rss-middleware
ENTRYPOINT ["./rss-torrent"]
