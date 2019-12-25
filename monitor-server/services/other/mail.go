package other

import (
	"net/smtp"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"strings"
	"fmt"
	"bytes"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
)

var (
	mailEnable bool
	smtpAuth smtp.Auth
	sendFrom string
	smtpAuthPassword  string
	smtpServer  string
	sender  string
)

func InitSmtpMail()  {
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
	smtpAuth = smtp.PlainAuth("",sendFrom,smtpAuthPassword,smtpServer)
}

func SendSmtpMail(smo m.SendAlertObj) {
	if !mailEnable {
		return
	}
	err := smtp.SendMail(fmt.Sprintf("%s:25", smtpServer), smtpAuth, sendFrom, smo.Accept, mailQQMessage(smo.Accept,smo.Subject,smo.Content))
	if err != nil {
		mid.LogError("send mail error", err)
	}
}

func mailQQMessage(to []string,subject,content string) []byte {
	var buff bytes.Buffer
	buff.WriteString("To:")
	buff.WriteString(strings.Join(to, ","))
	buff.WriteString("\r\nFrom:")
	buff.WriteString(sender+"<"+sendFrom+">")
	buff.WriteString("\r\nSubject:")
	buff.WriteString(subject)
	buff.WriteString("\r\nContent-Type:text/plain;charset=UTF-8\r\n\r\n")
	buff.WriteString(content)
	return buff.Bytes()
}