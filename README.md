# rss-middleware 
![Build](https://github.com/Bpazy/rss-middleware/workflows/Build/badge.svg)
![Docker Pulls](https://img.shields.io/docker/pulls/bpazy/rss-middleware)
  
rss-middleware 的目的是提取 RSS 订阅中的磁力，并将磁力推送到 qBittorrent 中。

## 使用手册
### 直接使用
```
Usage of rss-torrent.exe:
  -cron string
        守护模式
  -qbittorrent string
        qBittorrent API 链接地址
  -qbittorrent-password string
        qBittorrent 密码
  -qbittorrent-username string
        qBittorrent 用户名
  -rss string
        RSS 链接地址
```
### Docker 示例
```shell
docker run --name rss-torrent -e RSS=https://rsshub.app/dytt -e QBITTORRENT=http://192.168.194.20:8080 -e QBITTORRENT_USERNAME=admin -e QBITTORRENT_PASSWORD=admin -e CRON="*/1 * * * *" bpazy/rss-middleware
```

### Docker Compose
1. 下载 [docker-compose.yml](./docker-compose.yml) 到任意位置;
2. 编辑 docker-compose.yml 文件;
3. 启动: `docker-compose up -d`
