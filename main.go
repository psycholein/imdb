package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	imbdBase  = "http://www.imdb.com"
	imbdQuery = imbdBase + "/find?ref_=nv_sr_fn&q=%s&s=all"
)

func main() {
	dirs, err := ioutil.ReadDir("./")
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.OpenFile("__movies.txt", os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	for _, dir := range dirs {
		title := dir.Name()
		if strings.Index(title, ".") == -1 || strings.Index(title, ".txt") > -1 {
			continue
		}
		title = cleanTitle(title)
		query := fmt.Sprintf(imbdQuery, url.QueryEscape(title))
		name, path, rating := getInfo(query)
		fmt.Println(title, name, rating, path)

		movies := title + "\t" + name + "\t" + rating + "\t" + path + "\n"
		_, err := f.WriteString(movies)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func cleanTitle(str string) (name string) {
	name = str[0:strings.LastIndex(str, ".")]

	regex("\\[.*\\]", &name, "")
	regex("\\(.*\\)", &name, "")

	regex("\\bI\\b", &name, "1")
	regex("\\bII\\b", &name, "2")
	regex("\\bIII\\b", &name, "3")
	regex("\\bIV\\b", &name, "4")
	regex("\\bV\\b", &name, "5")
	regex("\\bVI\\b", &name, "6")
	regex("\\bVII\\b", &name, "7")
	regex("\\bVIII\\b", &name, "8")
	regex("\\bIX\\b", &name, "9")
	regex("\\bX\\b", &name, "10")
	regex("\\bXI\\b", &name, "11")
	regex("\\bXII\\b", &name, "12")
	regex("\\bXIII\\b", &name, "13")
	return
}

func regex(reg string, s *string, replace string) {
	re := regexp.MustCompile(reg)
	*s = re.ReplaceAllString(*s, replace)
}

func getInfo(query string) (name string, path string, rating string) {
	doc, err := goquery.NewDocument(query)
	if err != nil {
		log.Fatal(err)
	}
	doc.Find(".result_text a").First().Each(func(i int, s *goquery.Selection) {
		link, ok := s.Attr("href")
		if ok {
			path = imbdBase + link
			doc, err := goquery.NewDocument(path)
			if err != nil {
				log.Fatal(err)
			}
			doc.Find(".star-box-details strong span").First().Each(func(i int, s *goquery.Selection) {
				rating, _ = s.Html()
			})
			doc.Find("h1.header span").First().Each(func(i int, s *goquery.Selection) {
				name, _ = s.Html()
			})
		}
	})
	return
}
