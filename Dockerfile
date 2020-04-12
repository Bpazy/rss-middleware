FROM golang:latest AS development
RUN git clone --progress --verbose --depth=1 https://github.com/Bpazy/rss-middleware /rss-middleware
WORKDIR /rss-middleware
RUN go env && CGO_ENABLED=0 go build ./cmd/rss-torrent

FROM alpine:latest AS production
ENV CRON ""
ENV QBITTORRENT ""
ENV QBITTORRENT_PASSWORD ""
ENV QBITTORRENT_USERNAME ""
ENV RSS ""
COPY --from=development /rss-middleware/rss-torrent /rss-middleware/rss-torrent
WORKDIR /rss-middleware
ENTRYPOINT ./rss-torrent \
                -rss $RSS \
                -qbittorrent $QBITTORRENT \
                -qbittorrent-username $QBITTORRENT_USERNAME \
                -qbittorrent-password $QBITTORRENT_PASSWORD \
                -cron="$CRON"
