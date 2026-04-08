package send_mail

import (
	"fmt"

	mail "github.com/go-mail/mail/v2"
)

type Response struct {
	Error   bool                   `json:"error"`
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}
type InputMail struct {
	Subject string
	Content string
	Mail    string
	CC      []string
}

type Config struct {
	SENDER_MAIL string `json:"sender_mail"`
	ACCOUNT     string `json:"mail_account"`
	PASS        string `json:"mail_pass"`
	HeaderName  string `json:"header"`
	PORT        int    `json:"port"`
	HOST        string `json:"host"`
}

func SendMail(send_mail InputMail, config Config) error {
	var SENDER_MAIL = config.SENDER_MAIL
	var MAIL_ACCOUNT = config.ACCOUNT
	var PASS = config.PASS
	var HeaderName = config.HeaderName
	var PORT = config.PORT
	var HOST = config.HOST

	m := mail.NewMessage()
	m.SetHeader("From", HeaderName+"<"+SENDER_MAIL+">")
	m.SetHeader("To", send_mail.Mail)
	if len(send_mail.CC) > 0 {
		m.SetHeader("Cc", send_mail.CC...)
	}
	m.SetHeader("Subject", send_mail.Subject)
	m.SetBody("text/html", send_mail.Content)
	d := mail.NewDialer(HOST, PORT, MAIL_ACCOUNT, PASS)

	if err := d.DialAndSend(m); err != nil {
		fmt.Println("-------error send mail------->", err)
		return err
	}
	return nil
}
