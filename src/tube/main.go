package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type TubeVideoMetaDataRetriever struct {
	videoUrl   string
	videoBsoup *goquery.Document
}

func VideoMetaDataRetriever(videoUrl string) *TubeVideoMetaDataRetriever {
	videoBsoup, err := goquery.NewDocument(videoUrl)
	if err != nil {
		log.Fatal(err)
	}

	return &TubeVideoMetaDataRetriever{
		videoUrl:   videoUrl,
		videoBsoup: videoBsoup,
	}
}

func (t *TubeVideoMetaDataRetriever) metaContentTags() *goquery.Selection {
	return t.videoBsoup.Find("meta")
}

func main() {
	videoIds := []string{"GTWogFFA7TE", "BKLVpDTZOPQ", "qpfJRZfuesU"}

	// create directory for metadata file if it does not exist
	dir := "data/metadata/video"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatal(err)
		}
	}

	// create metadata file with suffix of current date in yyymmdd format
	filename := fmt.Sprintf("video_metadata_%s.json", time.Now().Format("20060102"))
	filepath := filepath.Join(dir, filename)

	jsonFile, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer jsonFile.Close()

	jsonData, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.Fatal(err)
	}

	metaData := make(map[string]map[string]string)
	if len(jsonData) > 0 {
		err = json.Unmarshal(jsonData, &metaData)
		if err != nil {
			log.Fatal(err)
		}
	}

	for _, videoId := range videoIds {
		if _, ok := metaData[videoId]; ok {
			// Video metadata already exists, skip
			fmt.Printf("Skipping video %s, metadata already exists in %s\n", videoId, filename)
			continue
		}

		videoUrl := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoId)
		t := VideoMetaDataRetriever(videoUrl)

		metaTags := t.metaContentTags()

		metaData[videoId] = make(map[string]string)
		metaTags.Each(func(i int, s *goquery.Selection) {
			itemProp := s.AttrOr("itemprop", "")
			metaContent := s.AttrOr("content", "")
			if itemProp != "" && metaContent != "" {
				metaData[videoId][itemProp] = metaContent
			}
		})

		jsonData, err := json.MarshalIndent(metaData, "", "    ")
		if err != nil {
			log.Fatal(err)
		}

		_, err = jsonFile.Seek(0, 0)
		if err != nil {
			log.Fatal(err)
		}

		_, err = jsonFile.Write(jsonData)
		if err != nil {
			log.Fatal(err)
		}

		err = jsonFile.Truncate(int64(len(jsonData)))
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Metadata for video %s saved to %s\n", videoId, filename)
	}
}
