package main

import (
	"fmt"
	"github.com/mmcdole/gofeed"
)

func main() {
	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL("https://rsshub-bva.herokuapp.com/eztv/torrents/6048596")
	fmt.Println(feed)
}
