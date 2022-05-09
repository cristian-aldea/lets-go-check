# lets-go-check

A configurable web crawler to check the status of your websites and APIs!

## Local Development

First, to run this project, you'll need to [install go](https://go.dev/doc/install).

Then, you'll need to download the project's dependencies:

```bash
go mod download
```

## Configuration

This application can be configured in many ways, using flags and environment variables.

### Flags

Available flags for this application are:
- `--config-path`: The path for the config file describing the tests to run
- `--email-to`: The email address to send alerts to

Run `go run main.go --help` for more details.

### Environment Variables

- SMTP_USERNAME: The user from which to send email alerts
- SMTP_PASSWORD: The password for the email account from which to send alert

## Deploy

Current a bash script is available for sending the necessary files to a server so the program can be run as a cronjob.

To deploy the application to a remote server:
```bash
bash ./deploy.sh . DESTINATION
```
Where:
- SOURCE: the project folder
- DESTINATION: The destination to deploy to