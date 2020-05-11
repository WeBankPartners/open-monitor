package other

import (
	"net/smtp"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"strings"
	"fmt"
	"bytes"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"crypto/tls"
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
	if m.Config().Alert.Mail.Tls {
		sendSMTPMailTLS(smo)
		return
	}
	err := smtp.SendMail(fmt.Sprintf("%s:25", smtpServer), smtpAuth, sendFrom, smo.Accept, mailQQMessage(smo.Accept,smo.Subject,smo.Content))
	if err != nil {
		mid.LogError("send mail error", err)
	}
}

func sendSMTPMailTLS(smo m.SendAlertObj)  {
	tlsConfig := &tls.Config{
		InsecureSkipVerify:true,
		ServerName: smtpServer,
	}
	address := fmt.Sprintf("%s:465", smtpServer)
	conn,err := tls.Dial("tcp", address, tlsConfig)
	if err != nil {
		mid.LogError("tls dial error ", err)
		return
	}
	client,err := smtp.NewClient(conn, smtpServer)
	if err != nil {
		mid.LogError("smtp new client error ", err)
		return
	}
	defer client.Close()
	if b,_ := client.Extension("AUTH"); b {
		err = client.Auth(smtpAuth)
		if err != nil {
			mid.LogError("client auth error ", err)
			return
		}
	}
	err = client.Mail(sendFrom)
	if err != nil {
		mid.LogError("client mail set from error ", err)
		return
	}
	for _,to := range smo.Accept {
		if err = client.Rcpt(to); err != nil {
			mid.LogError(fmt.Sprintf("client rcpt %s error ", to), err)
			return
		}
	}
	w,err := client.Data()
	if err != nil {
		mid.LogError("client data init error ", err)
		return
	}
	_,err = w.Write(mailQQMessage(smo.Accept, smo.Subject, smo.Content))
	if err != nil {
		mid.LogError("write message error ", err)
		return
	}
	w.Close()
	err = client.Quit()
	if err != nil {
		mid.LogError("client quit error ", err)
		return
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