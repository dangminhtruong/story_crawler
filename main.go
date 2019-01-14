package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"

	"./database"
	"./processXml"
	"github.com/gocolly/colly"
	"github.com/gosimple/slug"
)

func visitLink(urlSet processXml.Urlset, db *sql.DB) {
	for i := 0; i < len(urlSet.Urls); i++ {
		c := colly.NewCollector()
		c.OnHTML("#the-post", func(e *colly.HTMLElement) {
			title := e.ChildText("#the-post > div.post-head > h1")
			avataUrl := e.ChildAttr("#the-post > div.post-head > figure > div > img", "src")

			md5HashInBytes := md5.Sum([]byte(title))
			avata := hex.EncodeToString(md5HashInBytes[:])
			img, _ := os.Create("img/" + avata + ".jpg")
			defer img.Close()
			resp, _ := http.Get(avataUrl)
			defer resp.Body.Close()

			b, _ := io.Copy(img, resp.Body)
			fmt.Println("File size: ", b)

			content := ""

			e.ForEach("#the-post-content > p", func(_ int, m *colly.HTMLElement) {
				contentOrigin := regexp.MustCompile(`\n`)
				contentConverted := contentOrigin.ReplaceAllString(m.Text, "<br/>")
				content += "<p>" + contentConverted + "</p>"
			})

			slugName := slug.Make(title)

			insPost, err := db.Prepare("INSERT INTO dkn.posts(title, avata, content, slug) VALUES(?, ?, ?, ?)")
			if err != nil {
				fmt.Println("Insert failed")
			} else {
				insPost.Exec(title, "img/"+avata+".jpg", content, slugName)
				fmt.Printf("Inserted: %q\n", slugName)
			}
		})

		c.OnRequest(func(r *colly.Request) {
			fmt.Println("Visiting", r.URL)
		})

		c.Visit(urlSet.Urls[i].Loc)

	}
}

func main() {
	db := database.DBConn()
	defer db.Close()
	links := processXml.ReadSiteMap("sitemaptest.xml")
	visitLink(links, db)
}
