package main

import "net/smtp"

type mailTo struct {
	from        string
	password    string
	sendTo      []string
	smtpHost    string
	smtpPort    string
	mailMessage string
}

func (m mailTo) sendMail() error {
	messageByte := []byte(m.mailMessage)
	mailAuth := smtp.PlainAuth("", m.from, m.password, m.smtpHost)

	err := smtp.SendMail(m.smtpHost+":"+m.smtpPort, mailAuth, m.from, m.sendTo, messageByte)
	if err != nil {
		return err
	}

	return nil
}
