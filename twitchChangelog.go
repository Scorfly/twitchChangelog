// main.go

package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/rapidloop/skv"
)

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func main() {
	currentTime := time.Now()
	fmt.Println("Starting … ", currentTime.Format("2006-01-02 15:04:05"))

	/* get discord webhook url */
	discordWebhook := flag.String("discord", "127.0.0.1", "Discord webhook URL")
	flag.Parse()

	/***************************************************
	 *
	 *      GET CHANGELOG PAGE
	 *
	 */
	url := "https://dev.twitch.tv/docs/change-log"
	req, err := http.NewRequest("GET", url, nil)
	resp, _ := http.DefaultClient.Do(req)
	// handle the error if there is one
	if err != nil {
		panic(err)
	}
	// do this now so it won't be forgotten
	defer resp.Body.Close()
	// reads html as a slice of bytes
	respHtml, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(respHtml)))
	if err != nil {
		panic(err)
	}

	/***************************************************
	 *
	 *      PARSE CHANGELOG TO GET LAST UPDATE
	 *
	 */

	var cell []string
	var cells [][]string
	// Find each table
	doc.Find("table").Each(func(index int, tablehtml *goquery.Selection) {
		tablehtml.Find("tr").Each(func(indextr int, rowhtml *goquery.Selection) {
			rowhtml.Find("td").Each(func(indexth int, tablecell *goquery.Selection) {
				cell = append(cell, tablecell.Text())
			})
			cells = append(cells, cell)
			cell = nil
		})
	})

	changeLogDate := cells[1][0]
	changeLogDetails := cells[1][1]

	/***************************************************
	 *
	 *      Init storage and chech if key exists
	 *
	 */

	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// get last version form local db
	store, err := skv.Open(dir + "/.twitchChangelog.db")
	if err != nil {
		panic(err)
	}

	hashedContent := GetMD5Hash(changeLogDetails)

	var info string
	err = store.Get("last-twitch-version", &info)
	if err != nil {
		if err.Error() == "skv: key not found" {
			// init
			err = store.Put("last-twitch-version", "0")
			if err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}

	fmt.Println("Last stored version : ", info, " and changelog date is : ", changeLogDate)

	/* If last change already store, just exit */
	if info == changeLogDate {
		return
	}

	// check if value is already stored
	err = store.Get("stored::"+hashedContent, &info)
	if err != nil {
		if err.Error() == "skv: key not found" {
			// init
			err = store.Put("stored::"+hashedContent, "stored")
			if err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}

	if info == "stored" {
		return
	}

	/***************************************************
	 *
	 *      New version detected !
	 *          - store it
	 *          - post on discord the update
	 *
	 */

	err = store.Put("last-twitch-version", changeLogDate)
	if err != nil {
		panic(err)
	}

	changeLogDetails = strings.Replace(changeLogDetails, "\n", `\n`, -1)
	var jsonStr = []byte(`{"content": "**[` + changeLogDate + `]** \n\n` + changeLogDetails + `"}`)

	url = *discordWebhook
	req, err = http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("Updated …")
}
