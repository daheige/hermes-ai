package message

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"strings"
	"time"
)

type SMTPConfig struct {
	Server  string
	Port    int
	Account string
	From    string
	Token   string
}

func shouldAuth(cfg SMTPConfig) bool {
	return cfg.Account != "" || cfg.Token != ""
}

func SendEmail(cfg SMTPConfig, systemName, subject, receiver, content string) error {
	if receiver == "" {
		return fmt.Errorf("receiver is empty")
	}
	if cfg.From == "" { // for compatibility
		cfg.From = cfg.Account
	}

	encodedSubject := fmt.Sprintf("=?UTF-8?B?%s?=", base64.StdEncoding.EncodeToString([]byte(subject)))

	// Extract domain from SMTPFrom
	parts := strings.Split(cfg.From, "@")
	var domain string
	if len(parts) > 1 {
		domain = parts[1]
	}
	// Generate a unique Message-ID
	buf := make([]byte, 16)
	_, err := rand.Read(buf)
	if err != nil {
		return err
	}
	messageId := fmt.Sprintf("<%x@%s>", buf, domain)

	mail := []byte(fmt.Sprintf("To: %s\r\n"+
		"From: %s<%s>\r\n"+
		"Subject: %s\r\n"+
		"Message-ID: %s\r\n"+ // add Message-ID header to avoid being treated as spam, RFC 5322
		"Date: %s\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n\r\n%s\r\n",
		receiver, systemName, cfg.From, encodedSubject, messageId, time.Now().Format(time.RFC1123Z), content))

	auth := smtp.PlainAuth("", cfg.Account, cfg.Token, cfg.Server)
	addr := fmt.Sprintf("%s:%d", cfg.Server, cfg.Port)
	to := strings.Split(receiver, ";")

	if cfg.Port == 465 || !shouldAuth(cfg) {
		// need advanced client
		var conn net.Conn
		var err error
		if cfg.Port == 465 {
			tlsConfig := &tls.Config{
				InsecureSkipVerify: true,
				ServerName:         cfg.Server,
			}
			conn, err = tls.Dial("tcp", fmt.Sprintf("%s:%d", cfg.Server, cfg.Port), tlsConfig)
		} else {
			conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", cfg.Server, cfg.Port))
		}
		if err != nil {
			return err
		}
		client, err := smtp.NewClient(conn, cfg.Server)
		if err != nil {
			return err
		}
		defer client.Close()
		if shouldAuth(cfg) {
			if err = client.Auth(auth); err != nil {
				return err
			}
		}
		if err = client.Mail(cfg.From); err != nil {
			return err
		}
		receiverEmails := strings.Split(receiver, ";")
		for _, receiver := range receiverEmails {
			if err = client.Rcpt(receiver); err != nil {
				return err
			}
		}
		w, err := client.Data()
		if err != nil {
			return err
		}
		_, err = w.Write(mail)
		if err != nil {
			return err
		}
		err = w.Close()
		if err != nil {
			return err
		}
		return nil
	}
	err = smtp.SendMail(addr, auth, cfg.Account, to, mail)
	if err != nil && strings.Contains(err.Error(), "short response") { // 部分提供商返回该错误，但实际上邮件已经发送成功
		log.Printf("short response from SMTP server, return nil instead of error: %s", err.Error())
		return nil
	}

	return err
}
