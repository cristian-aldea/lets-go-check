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
	"path/filepath"
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

var executablePath, _ = os.Executable()
var executableDir = filepath.Dir(executablePath)
var logFilePath = executableDir + "/lets-go-check.log"

func main() {
	emailTo := flag.String("email-to", "", "Recipient for email alert when applications are down")
	configPath := flag.String("config-path", executableDir+"/checks.json", "Path to the file with the configuration for the checks to run")
	flag.Parse()

	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("main - ERROR - Unable to open file %s. %v", logFilePath, err)
	}
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))

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

	var checks []Check
	readConfig(*configPath, &checks)

	log.Println("main - Running checks")
	for _, check := range checks {
		for _, url := range check.Urls {
			resp, err := http.Get(url)
			if err != nil {
				processCheckFailure(fmt.Sprintf("Request failed for %s: %v", url, err), info)
			}

			if check.Code != resp.StatusCode {
				processCheckFailure(fmt.Sprintf("Wrong status code for %s: Expected %d, received %d", url, check.Code, resp.StatusCode), info)
			}

			rawBody, _ := ioutil.ReadAll(resp.Body)
			respBody := string(rawBody)

			if check.Body != "" && !strings.Contains(respBody, check.Body) {
				processCheckFailure(fmt.Sprintf("Response from %s doesn't contain expected content: \"%s\"", url, check.Body), info)
			}
		}

	}
	log.Println("main - All checks passed! Closing application.")
}

func processCheckFailure(message string, info smtpInfo) {
	sendAlert(message, info)
	log.Printf("processCheckFailure - ERROR - %s", message)
	log.Fatalf("processCheckFailure - ERROR - Alert has been sent. Exiting")
}

func sendAlert(body string, info smtpInfo) {
	auth := smtp.PlainAuth("", info.username, info.password, SmtpHost)
	message := "From: " + info.username + "\r\n" +
		"To: " + info.to + "\r\n" +
		"Subject: lets-go-check Alert\r\n\r\n" +
		body + "\r\n"

	err := smtp.SendMail(SmtpHost+":"+SmtpPort, auth, info.username, []string{info.to}, []byte(message))

	if err != nil {
		log.Fatalf("sendAlert - ERROR - failed to send email. %v", err)
	}
}

func readConfig(path string, checks *[]Check) {
	jsonFile, err := os.Open(path)

	if err != nil {
		log.Fatalf("readConfig - ERROR - Failed to open file under %s. %v", path, err)
	}
	defer func(jsonFile *os.File) {
		err := jsonFile.Close()
		if err != nil {
			log.Fatalf("readConfig - Failed to close file %s", path)
		}
	}(jsonFile)

	jsonRaw, err := ioutil.ReadAll(jsonFile)

	if err != nil {
		log.Fatalf("readConfig - ERROR - Failed to read file under %s. %v", path, err)
	}

	err = json.Unmarshal(jsonRaw, checks)
	if err != nil {
		log.Fatalf("readConfig - ERROR - Failed to parse JSON. %v", err)
	}
}
