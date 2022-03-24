package main

import (
	"encoding/json"
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

type smtpInfo struct {
	username string
	password string
	to       string
	host     string
	port     string
}

type Check struct {
	Name string   `json:"name"`
	Urls []string `json:"urls"`
	Code int      `json:"code"`
	Body string   `json:"body"`
}

const SmtpUsernameEnvName = "SMTP_USERNAME"
const SmtpPasswordEnvName = "SMTP_PASSWORD"
const EmailToFlagName = "email-to"
const SmtpHost = "smtp.gmail.com"
const SmtpPort = "587"
const LogFilePath = "/var/log/lets-go-check.log"

func main() {
	logFile, err := os.OpenFile(LogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("main - ERROR - Unable to open file %s. %v", LogFilePath, err)
	}

	log.SetOutput(io.MultiWriter(os.Stdout, logFile))

	emailTo := flag.String("email-to", "", "Recipient for email alert when applications are down")
	configPath := flag.String("config-path", "checks.json", "Path to the file with the configuration for the checks to run")
	flag.Parse()

	if *emailTo == "" {
		log.Fatalf("main - ERROR - The program flag --%s is required", EmailToFlagName)
	}

	smtpUsername := os.Getenv(SmtpUsernameEnvName)
	smtpPassword := os.Getenv(SmtpPasswordEnvName)

	if smtpUsername == "" || smtpPassword == "" {
		log.Fatalf("main - ERROR - Please specify the environment variables %s and %s", SmtpUsernameEnvName, SmtpPasswordEnvName)
	}

	info := smtpInfo{
		username: smtpUsername,
		password: smtpPassword,
		to:       *emailTo,
		host:     SmtpHost,
		port:     SmtpPort,
	}

	log.Printf("main - Reading checks configuration under %s", *configPath)
	var checks []Check

	readConfig(*configPath, &checks)

	log.Println("main - Starting Tests!")
	for _, check := range checks {
		log.Printf("main - %s: Test started", check.Name)

		for _, url := range check.Urls {
			resp, err := http.Get(url)
			if err != nil {
				processRequestFailure(fmt.Sprintf("URL: %s\nERROR: Request failed. %v", url, err), info)
			}

			if check.Code != resp.StatusCode {
				processRequestFailure(fmt.Sprintf("URL: %s\nERROR: Wrong status code (Expected: %d, Received: %d)", url, check.Code, resp.StatusCode), info)
			}

			rawBody, _ := ioutil.ReadAll(resp.Body)
			respBody := string(rawBody)

			if check.Body != "" && !strings.Contains(respBody, check.Body) {
				processRequestFailure(fmt.Sprintf("URL: %s\nERROR: Body doesn't contain expected content: \"%s\"", url, check.Body), info)
			}
		}
		log.Printf("main - %s - Test completed successfully", check.Name)
	}
}

func processRequestFailure(message string, info smtpInfo) {
	sendAlert(message, info)
	log.Fatalf("processRequestFailure - ERROR - %s", message)
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

func readConfig(path string, checks *[]Check) {
	jsonFile, err := os.Open(path)

	if err != nil {
		log.Fatalf("readConfig - ERROR - Failed to open file under %s. %v", path, err)
	}
	defer jsonFile.Close()

	jsonRaw, err := ioutil.ReadAll(jsonFile)

	if err != nil {
		log.Fatalf("readConfig - ERROR - Failed to read file under %s. %v", path, err)
	}

	err = json.Unmarshal(jsonRaw, checks)
	if err != nil {
		log.Fatalf("readConfig - ERROR - Failed to parse JSON. %v", err)
	}
}
