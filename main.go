package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"os"
	"time"
)

type train struct {
	Name    string
	InfoUrl string
	IsDelay bool
}

type discordImg struct {
	URL string `json:"url"`
	H   int    `json:"height"`
	W   int    `json:"width"`
}

type discordAuthor struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Icon string `json:"icon_url"`
}

type discordField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type discordEmbed struct {
	Title  string         `json:"title"`
	Desc   string         `json:"description"`
	URL    string         `json:"url"`
	Color  int            `json:"color"`
	Image  discordImg     `json:"image"`
	Author discordAuthor  `json:"author"`
	Fields []discordField `json:"fields"`
}

type discordHook struct {
	Username  string         `json:"username"`
	AvatarUrl string         `json:"avatar_url"`
	Content   string         `json:"content"`
	Embeds    []discordEmbed `json:"embeds"`
}

func getTrainsList() []*train {
	// Change this
	return []*train{
		{Name: "小田急小田原線", InfoUrl: "https://transit.yahoo.co.jp/diainfo/109/0"},
		{Name: "相模線", InfoUrl: "https://transit.yahoo.co.jp/diainfo/33/0"},
	}
}

func notifyToDiscord(message string, url string, level string) {
	var color = 0x666666
	var imageUrl = ""

	switch level {
	case "warning":
		color = 0xD50000
		imageUrl = "https://s3.sda1.net/firefish/contents/57230e34-ac4e-4670-a030-a23c03db35a4.jpg"
	case "info":
		color = 0x36B200
		imageUrl = "https://s3.sda1.net/firefish/contents/1bb7cd3d-71b5-4b06-8af4-975446d9d01e.jpg"
	}

	var notify discordHook
	discordUserID := os.Getenv("DISCORD_USERID")
	notify.Username = "遅延情報Bot"
	notify.Content = fmt.Sprintf("<@%s>", discordUserID)
	notify.Embeds = []discordEmbed{
		discordEmbed{
			Title:  "鉄道運行情報",
			Desc:   message,
			URL:    url,
			Color:  color,
			Author: discordAuthor{Name: "tarininfo-go"},
			Image:  discordImg{URL: imageUrl},
			Fields: []discordField{
				discordField{Name: "ソース", Value: "Yahoo", Inline: true},
			},
		},
	}

	postJson, _ := json.Marshal(notify)

	// discord webhook_url
	hookUrl := os.Getenv("DISCORD_WEBHOOK")
	res, err := http.Post(
		hookUrl,
		"application/json",
		bytes.NewBuffer(postJson),
	)

	if err != nil {
		log.Fatal("Failed to create notify!")
	}
	defer res.Body.Close()
}

func checkTrainStatus(trains []*train) {
	for i, t := range trains {
		client := &http.Client{}
		req, err := http.NewRequest("GET", t.InfoUrl, nil)
		if err != nil {
			log.Fatalln(err)
		}

		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/118.0")

		resp, err := client.Do(req)
		if err != nil {
			log.Fatalln(err)
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Fatal("Failed to scrape: server returned non-200 status code")
		}

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		// .Findでtroubleクラスのddタグを探す
		if doc.Find("dd.trouble").Length() > 0 {
			fmt.Printf("%sは遅延しています\n", t.Name)
			if !t.IsDelay {
				notifyToDiscord(fmt.Sprintf("%sが遅延しています。", t.Name), t.InfoUrl, "warning")
				trains[i].IsDelay = true
			}
		} else {
			fmt.Printf("%sは遅延していません\n", t.Name)
			if t.IsDelay {
				notifyToDiscord(fmt.Sprintf("%sの遅延は解消しました。", t.Name), t.InfoUrl, "info")
				trains[i].IsDelay = false
			}
		}

		// 3秒クールタイム
		time.Sleep(3 * time.Second)
	}
}

func main() {
	trains := getTrainsList()
	for {
		checkTrainStatus(trains)
		time.Sleep(40 * time.Second)
	}
}
