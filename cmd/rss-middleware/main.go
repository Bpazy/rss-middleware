package main

import (
	"encoding/json"
	"flag"
	"github.com/go-resty/resty/v2"
	"github.com/mmcdole/gofeed"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http/cookiejar"
	"os"
	"strings"
)

const (
	DatabaseFileName = "rss-middleware-database.json"
	Ok               = "Ok"
)

// RSS 中的磁力相关信息
type RssMagnet struct {
	Title  string // 磁力标题
	GUID   string // 磁力唯一ID
	Magnet string // 磁力链接
	Read   bool   // 是否已读
}

var (
	qBittorrentApiUrl   string
	qBittorrentUsername string
	qBittorrentPassword string
	daemonCron          string
	configPath          string
	rssUrl              string
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	flag.StringVar(&rssUrl, "rss", "", "RSS 链接地址")
	flag.StringVar(&qBittorrentApiUrl, "qbittorrent", "", "qBittorrent API 链接地址")
	flag.StringVar(&qBittorrentUsername, "qbittorrent-username", "", "qBittorrent 用户名")
	flag.StringVar(&qBittorrentPassword, "qbittorrent-password", "", "qBittorrent 密码")
	flag.StringVar(&daemonCron, "cron", "", "守护模式")
	flag.StringVar(&configPath, "config-path", ".", "配置文件、数据文件存储目录")
	flag.Parse()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		err := os.MkdirAll(configPath, 0777)
		if err != nil {
			log.Fatalf("创建配置文件目录失败: %+v", err)
		}
	}
}

func main() {
	if daemonCron == "" {
		downloadRSSOnce(rssUrl)
		return
	}

	c := cron.New()
	_, err := c.AddFunc(daemonCron, func() {
		downloadRSSOnce(rssUrl)
	})
	if err != nil {
		log.Fatalf("cron error: %+v", err)
	}
	log.Info("rss-middleware 守护模式启动成功")
	c.Run()
}

func downloadRSSOnce(rssUrl string) {
	log.Debugf("开始加载 RSS 链接: %s", rssUrl)

	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(rssUrl)
	if err != nil {
		log.Warnf("加载 RSS 链接失败: %+v", err)
		return
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
	client.SetHostURL(qBittorrentApiUrl + "/api/v2")

	// 设置 cookie jar
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	client.SetCookieJar(jar)

	// 登录
	loginResp, err := client.R().SetQueryParams(map[string]string{
		"username": qBittorrentUsername,
		"password": qBittorrentPassword,
	}).Get("/auth/login")
	if err != nil {
		log.Fatalf("%+v", err)
	}
	loginStatus := string(loginResp.Body())
	if !strings.Contains(loginStatus, Ok) {
		log.Warnf("登录失败：%s", loginStatus)
		return
	}

	// 查询所有已存储的 RSS 数据
	savedRssMagnet := queryAllRssMagnet()
	GUID2RssMagnet := map[string]RssMagnet{}
	for _, rssMagnet := range savedRssMagnet {
		GUID2RssMagnet[rssMagnet.GUID] = rssMagnet
	}

	// 是否有新增 RSS
	noneNewRss := true
	// 新建任务
	for _, magnet := range rssMagnets {
		// 判断该 RSS 是否已读
		if savedMagnet, ok := GUID2RssMagnet[magnet.GUID]; ok && savedMagnet.Read {
			continue
		}

		noneNewRss = false
		addTorrentResp, err := client.R().SetQueryParams(map[string]string{
			"urls": magnet.Magnet,
		}).Get("/torrents/add")
		if err != nil {
			log.Fatalf("%+v", err)
		}
		if strings.Contains(string(addTorrentResp.Body()), Ok) {
			log.Infof("添加磁力成功：%s", magnet.Magnet)
			magnet.Read = true
		} else {
			log.Infof("添加磁力失败：%s", magnet.Magnet)
		}
		savedRssMagnet = append(savedRssMagnet, magnet)
	}
	if noneNewRss {
		log.Debugf("无新增 RSS")
	}

	needSaveRssMagnetBytes, err := json.Marshal(savedRssMagnet)
	if err != nil {
		log.Fatalf("序列化 RSS 数据失败: %+v", err)
	}
	err = ioutil.WriteFile(getDatabaseFilePath(), needSaveRssMagnetBytes, 0777)
	if err != nil {
		log.Warnf("保存 RSS 数据失败: %+v", err)
	}
}

func queryAllRssMagnet() []RssMagnet {
	dataBytes, err := ioutil.ReadFile(getDatabaseFilePath())
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

func getDatabaseFilePath() string {
	return configPath + "/" + DatabaseFileName
}
