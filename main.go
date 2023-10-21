package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
)

type train struct {
	Name    string
	InfoUrl string
}

func getTrainsList() []train {
	// Change this
	return []train{
		{Name: "小田急小田原線", InfoUrl: "https://transit.yahoo.co.jp/diainfo/109/0"},
		{Name: "相模線", InfoUrl: "https://transit.yahoo.co.jp/diainfo/33/0"},
	}
}

func main() {
	for _, t := range getTrainsList() {
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

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		// .Findでtroubleクラスのddタグを探す
		if doc.Find("dd.trouble").Length() > 0 {
			fmt.Printf("%sは遅延しています\n", t.Name)
		} else {
			fmt.Printf("%sは遅延していません\n", t.Name)
		}
	}
}
