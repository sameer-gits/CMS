package main

import (
	"net/smtp"
	"os"
)

type MailTo struct {
	from        string
	password    string
	sendTo      []string
	smtpHost    string
	smtpPort    string
	mailMessage []byte
}

// this is using mailtrap now
func (m MailTo) sendMail() error {
	messageByte := []byte(m.mailMessage)
	mailAuth := smtp.PlainAuth("", os.Getenv("SMTP_USERNAME"), m.password, m.smtpHost)

	err := smtp.SendMail(m.smtpHost+":"+m.smtpPort, mailAuth, m.from, m.sendTo, messageByte)
	if err != nil {
		return err
	}

	return nil
}
