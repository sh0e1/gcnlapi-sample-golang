package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
)

var urls = []string{
	"https://language.googleapis.com/v1/documents:analyzeEntities",
	"https://language.googleapis.com/v1/documents:analyzeSyntax",
}
var wg sync.WaitGroup
var key *string

type (
	document struct {
		Type     string `json:"type"`
		Language string `json:"language"`
		Content  string `json:"content"`
	}
	requestBody struct {
		Document     document `json:"document"`
		EncodingType string   `json:"encodingType"`
	}
)

func newRequestBody(s string) *requestBody {
	return &requestBody{
		Document: document{
			Type:     "PLAIN_TEXT",
			Language: "ja",
			Content:  s,
		},
		EncodingType: "UTF8",
	}
}

func main() {
	key = flag.String("key", os.Getenv("API_KEY"), "API Key")
	flag.Parse()

	if *key == "" {
		fmt.Print(`API Key is not set. You have to:
export API_KEY=YOUR_API_KEY or go run main.go -key=YOUR_API_KEY
`)
		os.Exit(1)
	}

	s := bufio.NewScanner(os.Stdin)
	client := &http.Client{}
	for s.Scan() {
		for _, url := range urls {
			wg.Add(1)
			go func(url string) {
				defer wg.Done()

				body, err := json.Marshal(newRequestBody(s.Text()))
				if err != nil {
					log.Fatalln(err)
					return
				}

				req, err := http.NewRequest(
					"POST",
					fmt.Sprintf("%s?key=%s", url, *key),
					bytes.NewBuffer(body),
				)
				if err != nil {
					log.Fatalln(err)
					return
				}

				resp, err := client.Do(req)
				if err != nil {
					log.Fatalln(err)
					return
				}
				defer resp.Body.Close()

				result, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Fatalln(err)
					return
				}
				log.Println(string(result))
			}(url)
		}
		wg.Wait()
	}
}
