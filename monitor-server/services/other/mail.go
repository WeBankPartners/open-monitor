package other

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"go.uber.org/zap"
	"net/smtp"
	"strings"
)

var (
	mailEnable       bool
	smtpAuth         smtp.Auth
	sendFrom         string
	smtpAuthPassword string
	smtpServer       string
	sender           string
)

func InitSmtpMail() {
	mailEnable = true
	if !m.Config().Alert.Mail.Enable {
		mailEnable = false
		return
	}
	if m.Config().Alert.Mail.Protocol != "smtp" {
		mailEnable = false
		return
	}
	sendFrom = m.Config().Alert.Mail.User
	smtpAuthPassword = m.Config().Alert.Mail.Password
	smtpServer = m.Config().Alert.Mail.Server
	sender = m.Config().Alert.Mail.Sender
	smtpAuth = smtp.PlainAuth("", sendFrom, smtpAuthPassword, smtpServer)
}

func SendSmtpMail(smo m.SendAlertObj) {
	if !mailEnable {
		return
	}
	if m.Config().Alert.Mail.Tls {
		sendSMTPMailTLS(smo)
		return
	}
	err := smtp.SendMail(fmt.Sprintf("%s:25", smtpServer), smtpAuth, sendFrom, smo.Accept, mailQQMessage(smo.Accept, smo.Subject, smo.Content))
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Send mail error", zap.Error(err))
	}
}

func sendSMTPMailTLS(smo m.SendAlertObj) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         smtpServer,
	}
	address := fmt.Sprintf("%s:465", smtpServer)
	conn, err := tls.Dial("tcp", address, tlsConfig)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Tls dial error", zap.Error(err))
		return
	}
	client, err := smtp.NewClient(conn, smtpServer)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Smtp new client error", zap.Error(err))
		return
	}
	defer client.Close()
	if b, _ := client.Extension("AUTH"); b {
		err = client.Auth(smtpAuth)
		if err != nil {
			log.Error(nil, log.LOGGER_APP, "Client auth error", zap.Error(err))
			return
		}
	}
	err = client.Mail(sendFrom)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Client mail set from error", zap.Error(err))
		return
	}
	for _, to := range smo.Accept {
		if err = client.Rcpt(to); err != nil {
			log.Error(nil, log.LOGGER_APP, fmt.Sprintf("Client rcpt %s error", to), zap.Error(err))
			return
		}
	}
	w, err := client.Data()
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Client data init error", zap.Error(err))
		return
	}
	_, err = w.Write(mailQQMessage(smo.Accept, smo.Subject, smo.Content))
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Write message error", zap.Error(err))
		return
	}
	w.Close()
	err = client.Quit()
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Client quit error", zap.Error(err))
		return
	}
}

func mailQQMessage(to []string, subject, content string) []byte {
	var buff bytes.Buffer
	buff.WriteString("To:")
	buff.WriteString(strings.Join(to, ","))
	buff.WriteString("\r\nFrom:")
	buff.WriteString(sender + "<" + sendFrom + ">")
	buff.WriteString("\r\nSubject:")
	buff.WriteString(subject)
	buff.WriteString("\r\nContent-Type:text/plain;charset=UTF-8\r\n\r\n")
	buff.WriteString(content)
	return buff.Bytes()
}
