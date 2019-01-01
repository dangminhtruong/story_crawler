package main

import (
	"net/http"
	"io"
	"os"
	"github.com/gocolly/colly"
	"fmt"
	"./processXml"
	"regexp"
	"./database"
	"database/sql"
	"strings"
	"github.com/gosimple/slug"
	"time"
	"crypto/md5"
	"encoding/hex"
	"flag"
)

func visitLink(urlSet processXml.Urlset, db *sql.DB, cate string, id int, path string) {
	for i := 0; i < len(urlSet.Urls); i++ {

		var urlCheck = regexp.MustCompile(cate + ".*[^0-9]$")
		var isVideo = regexp.MustCompile("video")
		var isTag = regexp.MustCompile("tag")
		if urlCheck.MatchString(urlSet.Urls[i].Loc) && ! isVideo.MatchString(urlSet.Urls[i].Loc) && !isTag.MatchString(urlSet.Urls[i].Loc){
			fmt.Println("Matched")
			c := colly.NewCollector()
			c.OnHTML(".post", func(e *colly.HTMLElement) {
				title := e.ChildText("header > h1")
				avataUrl := e.ChildAttr("header > center > img", "src")
				md5HashInBytes := md5.Sum([]byte(title))
				avata := hex.EncodeToString(md5HashInBytes[:])
				img, _ := os.Create(path + avata + ".jpg")
				defer img.Close()
				resp, _ := http.Get(avataUrl)
    			defer resp.Body.Close()

				b, _ := io.Copy(img, resp.Body)
				fmt.Println("File size: ", b)

				content := ""
				choseDesc :=  e.ChildText("div > p:nth-child(1)")
				choseDescSplited := strings.SplitAfter(choseDesc, ".")

				//fmt.Println(choseDescSplited) #post-1953 > header > center > img

				e.ForEach("div > p", func(_ int, m *colly.HTMLElement) {
					contentOrigin := regexp.MustCompile(`\n`)
					contentConverted := contentOrigin.ReplaceAllString(m.Text, "<br/>")
					content += "<p>" + contentConverted + "</p>"
				})

				slugName := slug.Make(title)
				var short_descript string
				if len(choseDescSplited) > 1{
					short_descript = choseDescSplited[0] + choseDescSplited[1]
				}else{
					short_descript = choseDescSplited[0]
				}
				
				share_link := "https://truyencotich.top/doc-truyen/" + slugName
				meta_keyword := title + ", truyện cổ tích, đọc truyện cổ tích, truyện cổ tích việt nam, truyện cổ tihs thế giới, cổ tích, cổ tích truyện, truyện cổ grin, truyện cổ tích nước ngoài"
				insPost, err := db.Prepare("INSERT INTO posts(title, short_descript, avata, content, slug, total_view, status, share_link, meta_title, meta_descript, meta_keyword, category_id, created_at)" + 
											" VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",)
				if err != nil {
					fmt.Println("Insert failed")
				}
				
				insPost.Exec(title, short_descript, "img/" + avata + ".jpg", content, slugName, 0, 1, share_link, title, short_descript, meta_keyword, id, time.Now()) 
				fmt.Printf("Inserted: %q\n", title)
			})

			c.OnRequest(func(r *colly.Request) {
				fmt.Println("Visiting", r.URL)
			})
		
			c.Visit(urlSet.Urls[i].Loc)
		}else{
			fmt.Println("Not match")
		}
	}
}

func main() {
	var cate string
	var path string
	var id int

	flag.StringVar(&cate, "cate", "", "category")
	flag.StringVar(&path, "path", "", "img store path example: /var/www/truyencotich.top/storage/app/public/img/")
	flag.IntVar(&id, "id", 0, "category id")
	flag.Parse()

	db := database.DBConn()
	defer db.Close()

	links := processXml.ReadSiteMap("sitemap.xml")
	if cate != "cate" && id > 0{
		visitLink(links, db, cate, id, path)
	}else{
		fmt.Println("Not enough arguments !")
	}
	
}


