package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type Test struct {
	url  string
	code int
	body string
}

var tests = []Test{
	{url: "https://www.cristianaldea.com", code: 200, body: "app-content"},
	{url: "https://cristianaldea.com", code: 200, body: "app-content"},
}

func main() {
	log.Println("Testing")

	for _, test := range tests {
		log.Printf("main - Testing %s", test.url)
		resp, err := http.Get(test.url)
		if err != nil {
			log.Fatalf("main - ERROR - %s: Failed to request url")
		}

		if test.code != resp.StatusCode {
			log.Fatalf("main - ERROR - %s: Wrong status code (Expected: %d, Received: %d)", test.url, test.code, resp.StatusCode)
		}

		rawBody, _ := ioutil.ReadAll(resp.Body)
		respBody := string(rawBody)

		if !strings.Contains(respBody, test.body) {
			log.Fatalf("main - ERROR - %s: Body doesn't contain \"%s\"", test.url, test.body)
		}
	}
}
