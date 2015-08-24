package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/PuerkitoBio/goquery"
	"github.com/shenwei356/util/bytesize"
	"gopkg.in/yaml.v2"
)

const (
	imdbBase  = "http://www.imdb.com"
	imdbQuery = imdbBase + "/find?ref_=nv_sr_fn&q=%s&s=all"

	imdbSection     = ".findSection"
	imdbSectionName = "Titles"
	imdbResult      = ".result_text"
	imdbMovie       = "h1.header span"
	imdbRating      = ".star-box-details strong span"
	imdbUsers       = ".star-box-details a span"
	imdbFSK         = ".infobar > meta"
	imdbDuration    = ".infobar > time"

	imdbSeries = "TV Series"
)

type Movie struct {
	Movie, Duration, Rating, Users, Fsk, Link, File, Size string
}

func main() {
	f, err := os.Create("_movies.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	h, err := os.Create("_movies.html")
	if err != nil {
		log.Fatal(err)
	}
	defer h.Close()

	_, err = h.WriteString(strings.TrimSpace(HtmlStart))
	if err != nil {
		log.Fatal(err)
	}

	table := template.Must(template.New("table").Parse(HtmlTable))

	var dirs []string
	readYaml("_dirs.yml", &dirs)
	fmt.Println(dirs)

	for _, dir := range dirs {
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			log.Fatal(err)
		}
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			title := file.Name()
			if strings.Index(title, ".") == -1 || strings.Index(title, "_") == 0 {
				continue
			}

			path := dir + file.Name()
			title = cleanTitle(title)
			query := fmt.Sprintf(imdbQuery, url.QueryEscape(title))
			doc, link, ok := getResult(query)
			if !ok {
				fmt.Println(dir + file.Name())

				m := Movie{File: path}
				err = table.Execute(h, m)
				if err != nil {
					log.Fatal(err)
				}

				_, err = f.WriteString(path + "\n")
				if err != nil {
					log.Fatal(err)
				}
				continue
			}

			movie := getInfo(doc, imdbMovie)
			rating := getInfo(doc, imdbRating)
			users := getInfo(doc, imdbUsers)
			fsk := getInfoAttr(doc, imdbFSK, "content")
			duration := getInfo(doc, imdbDuration)
			size := bytesize.ByteSize(file.Size()).String()

			fmt.Println(path, movie, rating, users, fsk, duration, link, size)

			m := Movie{File: path, Movie: movie, Duration: duration, Size: size,
				Rating: rating, Fsk: fsk, Link: link, Users: users}
			err = table.Execute(h, m)
			if err != nil {
				log.Fatal(err)
			}

			movies := path + "\t" + movie + "\t" + rating + "\t" + users + "\t"
			movies += fsk + "\t" + duration + "\t" + size + "\t" + link + "\n"
			_, err = f.WriteString(movies)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	_, err = h.WriteString(strings.TrimSpace(HtmlEnd))
	if err != nil {
		log.Fatal(err)
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

func getResult(query string) (doc *goquery.Document, link string, found bool) {
	doc, err := goquery.NewDocument(query)
	if err != nil {
		log.Fatal(err)
	}

	var ok bool
	found = false
	doc.Find(imdbSection).Each(func(i int, s *goquery.Selection) {
		section, err := s.Html()
		if err != nil {
			return
		}
		if strings.Index(section, imdbSectionName) == -1 {
			return
		}
		s.Find(imdbResult).Each(func(i int, s *goquery.Selection) {
			if found {
				return
			}
			content, _ := s.Html()
			if strings.Index(content, imdbSeries) != -1 {
				return
			}
			s.Find("a").First().Each(func(i int, s *goquery.Selection) {
				link, ok = s.Attr("href")
				if ok && strings.Index(link, "/title/") != -1 {
					found = true
				}
			})
		})
	})
	if found {
		link = imdbBase + link
		doc = getMoviePage(link)
	}
	return
}

func getMoviePage(link string) (doc *goquery.Document) {
	var err error
	doc, err = goquery.NewDocument(link)
	if err != nil {
		log.Fatal(err)
	}
	return
}

func getInfo(doc *goquery.Document, query string) (r string) {
	doc.Find(query).First().Each(func(i int, s *goquery.Selection) {
		r, _ = s.Html()
	})
	r = strings.TrimSpace(r)
	return
}

func getInfoAttr(doc *goquery.Document, query string, attr string) (r string) {
	doc.Find(query).First().Each(func(i int, s *goquery.Selection) {
		r, _ = s.Attr(attr)
	})
	r = strings.TrimSpace(r)
	return
}

func readYaml(filename string, data interface{}) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	err = yaml.Unmarshal(b, data)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}
