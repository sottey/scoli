package api

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
)

func sendEmail(smtpSettings EmailSMTPSettings, subject, body string) error {
	from := mail.Address{Address: smtpSettings.From}
	to := mail.Address{Address: smtpSettings.To}
	headers := map[string]string{
		"From":         from.String(),
		"To":           to.String(),
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/plain; charset=UTF-8",
	}

	var msg bytes.Buffer
	for key, value := range headers {
		_, _ = msg.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}
	_, _ = msg.WriteString("\r\n")
	_, _ = msg.WriteString(body)

	addr := net.JoinHostPort(smtpSettings.Host, fmt.Sprintf("%d", smtpSettings.Port))
	auth := smtp.PlainAuth("", smtpSettings.Username, smtpSettings.Password, smtpSettings.Host)

	if smtpSettings.UseTLS && smtpSettings.Port == 465 {
		return sendWithTLS(addr, smtpSettings.Host, auth, from.Address, to.Address, msg.Bytes())
	}
	return sendWithStartTLS(addr, smtpSettings.Host, auth, smtpSettings.UseTLS, from.Address, to.Address, msg.Bytes())
}

func sendWithTLS(addr, host string, auth smtp.Auth, from, to string, msg []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: host})
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()

	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return err
		}
	}
	if err := client.Mail(from); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := writer.Write(msg); err != nil {
		_ = writer.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return client.Quit()
}

func sendWithStartTLS(addr, host string, auth smtp.Auth, useTLS bool, from, to string, msg []byte) error {
	client, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer client.Close()

	if useTLS {
		if err := client.StartTLS(&tls.Config{ServerName: host}); err != nil {
			return err
		}
	}
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return err
		}
	}
	if err := client.Mail(from); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := writer.Write(msg); err != nil {
		_ = writer.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return client.Quit()
}
