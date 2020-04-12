FROM golang:latest AS development
RUN git clone --progress --verbose --depth=1 https://github.com/Bpazy/rss-middleware /rss-middleware
WORKDIR /rss-middleware
RUN go env && CGO_ENABLED=0 go build ./cmd/rss-middleware

FROM alpine:latest AS production
ENV CRON ""
ENV QBITTORRENT ""
ENV QBITTORRENT_PASSWORD ""
ENV QBITTORRENT_USERNAME ""
ENV RSS ""
VOLUME /data
COPY --from=development /rss-middleware/rss-middleware /rss-middleware/rss-middleware
WORKDIR /rss-middleware
ENTRYPOINT ./rss-middleware \
                -rss $RSS \
                -qbittorrent $QBITTORRENT \
                -qbittorrent-username $QBITTORRENT_USERNAME \
                -qbittorrent-password $QBITTORRENT_PASSWORD \
                -config-path /data \
                -cron="$CRON"
