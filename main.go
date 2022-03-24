package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"
)

type Test struct {
	name string
	urls []string
	code int
	body string
}

type smtpInfo struct {
	username string
	password string
	to       string
	host     string
	port     string
}

const SmtpUsernameEnvName = "SMTP_USERNAME"
const SmtpPasswordEnvName = "SMTP_PASSWORD"
const EmailToFlagName = "email-to"
const SmtpHost = "smtp.gmail.com"
const SmtpPort = "587"
const LogFilePath = "/var/log/lets-go-check.log"

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
	logFile, err := os.OpenFile(LogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("main - ERROR: Unable to open file %s. %v", LogFilePath, err)
	}

	log.SetOutput(io.MultiWriter(os.Stdout, logFile))

	emailTo := flag.String("email-to", "", "Recipient for email alert when applications are down")
	flag.Parse()

	if *emailTo == "" {
		log.Fatalf("main - ERROR: The program flag --%s is required", EmailToFlagName)
	}

	smtpUsername := os.Getenv(SmtpUsernameEnvName)
	smtpPassword := os.Getenv(SmtpPasswordEnvName)

	if smtpUsername == "" || smtpPassword == "" {
		log.Fatalf("main - ERROR: Please specify the environment variables %s and %s", SmtpUsernameEnvName, SmtpPasswordEnvName)
	}

	info := smtpInfo{
		username: smtpUsername,
		password: smtpPassword,
		to:       *emailTo,
		host:     SmtpHost,
		port:     SmtpPort,
	}

	log.Println("main: Starting Tests!")
	for _, test := range tests {
		log.Printf("main - %s: Test started", test.name)

		for _, url := range test.urls {
			resp, err := http.Get(url)
			if err != nil {
				processRequestFailure(fmt.Sprintf("URL: %s\nERROR: Request failed. Details: %v", url, err), info)
			}

			if test.code != resp.StatusCode {
				processRequestFailure(fmt.Sprintf("URL: %s\nERROR: Wrong status code (Expected: %d, Received: %d)", url, test.code, resp.StatusCode), info)
			}

			rawBody, _ := ioutil.ReadAll(resp.Body)
			respBody := string(rawBody)

			if test.body != "" && !strings.Contains(respBody, test.body) {
				processRequestFailure(fmt.Sprintf("URL: %s\nERROR: Body doesn't contain expected content: \"%s\"", url, test.body), info)
				log.Fatalf("main - %s - ERROR - Body doesn't contain expected content: \"%s\"", test.name, test.body)
			}
		}

		log.Printf("main - %s - Test completed successfully", test.name)
	}
}

func processRequestFailure(message string, info smtpInfo) {
	log.Printf("processRequestFailure - ERROR - %s. Sending alert", message)
	sendAlert(message, info)
	os.Exit(1)
}

func sendAlert(body string, info smtpInfo) {
	auth := smtp.PlainAuth("", info.username, info.password, SmtpHost)
	message := "From: " + info.username + "\r\n" +
		"To: " + info.to + "\r\n" +
		"Subject: lets-go-check Alert\r\n\r\n" +
		body + "\r\n"

	log.Printf("sendAlert - Sending alert to %s", info.to)
	err := smtp.SendMail(SmtpHost+":"+SmtpPort, auth, info.username, []string{info.to}, []byte(message))

	if err != nil {
		log.Printf("sendAlert - ERROR - %v", err)
	} else {
		log.Print("sendAlert - Email sent successfully!")
	}
}
