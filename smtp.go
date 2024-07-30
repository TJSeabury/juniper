package main

import (
	"net/smtp"
	"strconv"
)

type Email struct {
	To      []string
	From    string
	Subject string
	Body    string
}

type Emailer interface {
	Send(email Email) error
	Initialize(username, password string, host string) error
}

type Mailer struct {
	Host     string
	Port     int // Common ports are 25, 465 (SSL required), 587 (TLS required)
	Username string
	Password string
	Auth     smtp.Auth
}

func (m *Mailer) Initialize(username, password, host string) {
	m.Auth = smtp.PlainAuth("", username, password, host)
}

func (m *Mailer) Send(email Email) error {
	err := smtp.SendMail(
		m.Host+":"+strconv.Itoa(m.Port),
		m.Auth,
		email.From,
		email.To,
		[]byte(email.Body),
	)
	if err != nil {
		return err
	}

	return nil
}
