package main

import (
	"encoding/json"
	"flag"
	"github.com/go-resty/resty/v2"
	"github.com/mmcdole/gofeed"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http/cookiejar"
	"os"
	"strings"
)

const (
	DatabaseFileName = "rss-middleware-database.json"
)

// RSS 中的磁力相关信息
type RssMagnet struct {
	Title  string // 磁力标题
	GUID   string // 磁力唯一ID
	Magnet string // 磁力链接
	Read   bool   // 是否已读
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

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
	var rssMagnets []RssMagnet
	for _, item := range feed.Items {
		for _, enclosure := range item.Enclosures {
			rssMagnets = append(rssMagnets, RssMagnet{
				Title:  item.Title,
				GUID:   item.GUID,
				Magnet: enclosure.URL,
				Read:   false,
			})
		}
	}

	var magnets []string
	for _, rssMagnet := range rssMagnets {
		magnets = append(magnets, rssMagnet.Magnet)
	}
	log.Debugf("获取到磁力链接：%s", magnets)

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
	log.Debugf("登录状态：%s\n", string(loginResp.Body()))

	// 查询所有已存储的 RSS 数据
	savedRssMagnet := queryAllRssMagnet()
	GUID2RssMagnet := map[string]RssMagnet{}
	for _, rssMagnet := range savedRssMagnet {
		GUID2RssMagnet[rssMagnet.GUID] = rssMagnet
	}

	// 新建任务
	for _, magnet := range rssMagnets {
		// 判断该 RSS 是否已读
		if savedMagnet, ok := GUID2RssMagnet[magnet.GUID]; ok && savedMagnet.Read {
			continue
		}

		addTorrentResp, err := client.R().SetQueryParams(map[string]string{
			"urls": magnet.Magnet,
		}).Get("/torrents/add")
		if err != nil {
			log.Fatalf("%+v\n", err)
		}
		if strings.Contains(string(addTorrentResp.Body()), "Ok") {
			log.Infof("添加磁力成功：%s\n", magnet.Magnet)
			magnet.Read = true
		} else {
			log.Infof("添加磁力失败：%s\n", magnet.Magnet)
		}
		savedRssMagnet = append(savedRssMagnet, magnet)
	}

	needSaveRssMagnetBytes, err := json.Marshal(savedRssMagnet)
	if err != nil {
		log.Fatalf("序列化 RSS 数据失败: %+v\n", err)
	}
	err = ioutil.WriteFile(DatabaseFileName, needSaveRssMagnetBytes, 0777)
	if err != nil {
		log.Warnf("保存 RSS 数据失败: %+v\n", err)
	}
}

func queryAllRssMagnet() []RssMagnet {
	dataBytes, err := ioutil.ReadFile(DatabaseFileName)
	if err != nil {
		log.Warnf("查询既存数据失败：%+v", err)
		return nil
	}
	var result []RssMagnet
	err = json.Unmarshal(dataBytes, &result)
	if err != nil {
		log.Warnf("序列化失败：%+v", err)
		return nil
	}
	return result
}
