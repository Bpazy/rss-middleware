package main

import (
	"flag"
	"github.com/go-resty/resty/v2"
	"github.com/mmcdole/gofeed"
	"log"
	"net/http/cookiejar"
	"strings"
)

func main() {
	rssUrl := flag.String("rss", "", "rss url")
	qBittorrentApiUrl := flag.String("qbittorrent", "", "qbittorrent api url")
	qBittorrentUsername := flag.String("qbittorrent-username", "", "qbittorrent username")
	qBittorrentPassword := flag.String("qbittorrent-password", "", "qbittorrent password")
	flag.Parse()

	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(*rssUrl)
	if err != nil {
		log.Fatalf("%+v\n", err)
	}

	// 提取所有磁力
	var magnets []string
	for _, item := range feed.Items {
		for _, enclosure := range item.Enclosures {
			magnets = append(magnets, enclosure.URL)
		}
	}
	log.Printf("获取到磁力链接：%s", magnets)

	client := resty.New()
	client.SetHostURL(*qBittorrentApiUrl + "/api/v2")

	// 设置 cookie jar
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalf("%+v\n", err)
	}
	client.SetCookieJar(jar)

	// 登录
	loginResp, err := client.R().SetQueryParams(map[string]string{
		"username": *qBittorrentUsername,
		"password": *qBittorrentPassword,
	}).Get("/auth/login")
	if err != nil {
		log.Fatalf("%+v\n", err)
	}
	log.Printf("登录状态：%s\n", string(loginResp.Body()))

	// 新建任务
	for _, magnet := range magnets {
		addTorrentResp, err := client.R().SetQueryParams(map[string]string{
			"urls": magnet,
		}).Get("/torrents/add")
		if err != nil {
			log.Fatalf("%+v\n", err)
		}
		if strings.Contains(string(addTorrentResp.Body()), "Ok") {
			log.Printf("添加磁力成功：%s\n", magnet)
		} else {
			log.Printf("添加磁力失败：%s\n", magnet)
		}
	}
}
