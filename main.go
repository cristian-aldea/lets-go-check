package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type Test struct {
	name string
	urls []string
	code int
	body string
}

var tests = []Test{
	{
		name: "Main website",
		urls: []string{
			"https://www.cristianaldea.com",
			"https://cristianaldea.com",
			"http://www.cristianaldea.com",
		},
		code: 200,
		body: "Today is a gift!",
	},
	{
		name: "Starlog",
		urls: []string{
			"https://blog.cristianaldea.com",
			"http://blog.cristianaldea.com",
		},
		code: 200,
		body: "Welcome to ⭐ Starlog ⭐",
	},
	{
		name: "VoteIt",
		urls: []string{
			"https://voteit.cristianaldea.com",
			"http://voteit.cristianaldea.com",
			"https://voteit.cristianaldea.com/create-poll",
		},
		code: 200,
		body: "The place to create poll easily, quickly, and with no signup!",
	},
	{
		name: "VoteIt API",
		urls: []string{
			"https://api.voteit.cristianaldea.com/api/healthz",
		},
		code: 204,
		body: "",
	},
	{
		name: "Mull Recognition",
		urls: []string{
			"https://mull.cristianaldea.com",
			"http://mull.cristianaldea.com",
		},
		code: 200,
		body: "A standalone application for detecting waste!",
	},
}

func main() {
	log.Println("Testing")

	for _, test := range tests {
		log.Printf("main - %s - Test started", test.name)

		for _, url := range test.urls {
			resp, err := http.Get(url)
			if err != nil {
				log.Fatalf("main - %s - ERROR: Failed to request url", test.name)
			}

			if test.code != resp.StatusCode {
				log.Fatalf("main - %s - ERROR: Wrong status code (Expected: %d, Received: %d)", test.name, test.code, resp.StatusCode)
			}

			rawBody, _ := ioutil.ReadAll(resp.Body)
			respBody := string(rawBody)

			if test.body != "" && !strings.Contains(respBody, test.body) {
				log.Fatalf("main - %s - ERROR: Body doesn't contain expected content: \"%s\"", test.name, test.body)
			}
		}

		log.Printf("main - %s - Test completed successfully", test.name)

	}
}
