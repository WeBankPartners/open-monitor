package smtp

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"regexp"
	"strings"
)

type MailSender struct {
	SenderName   string
	SenderMail   string
	AuthUser     string
	AuthServer   string
	AuthPassword string
	SSL          bool
	Auth         Auth
	ByStartTLS   bool
}

func (ms *MailSender) Init() error {
	if ms.AuthServer == "" || ms.SenderMail == "" {
		return fmt.Errorf("Mail server and sender can not empty ")
	}
	if !VerifyMailAddress(ms.SenderMail) {
		return fmt.Errorf("Sender mail:%s validate fail ", ms.SenderMail)
	}
	if !strings.Contains(ms.AuthServer, ":") {
		if ms.SSL {
			ms.AuthServer = ms.AuthServer + ":465"
		} else {
			ms.AuthServer = ms.AuthServer + ":25"
		}
	}
	if ms.AuthUser == "" {
		ms.AuthUser = ms.SenderMail
	}
	ms.Auth = PlainAuth("", ms.AuthUser, ms.AuthPassword, ms.AuthServer)
	return nil
}

func (ms *MailSender) Send(subject, content string, addressee []string) error {
	var err error
	if subject == "" {
		return fmt.Errorf("Mail subject can not empty ")
	}
	if len(addressee) == 0 {
		return fmt.Errorf("Mail addressee can not empty ")
	}
	for _, to := range addressee {
		if !VerifyMailAddress(to) {
			err = fmt.Errorf("Mail:%s validate fail ", to)
			break
		}
	}
	if err != nil {
		return err
	}
	if ms.SSL {
		if ms.ByStartTLS {
			err = ms.sendStartTLSMail(subject, content, addressee)
		} else {
			err = ms.sendTLSMail(subject, content, addressee)
		}
	} else {
		err = SendMail(ms.AuthServer, ms.Auth, ms.SenderMail, addressee, mailQQMessage(addressee, subject, content, ms.SenderName, ms.SenderMail))
	}
	return err
}

func (ms *MailSender) sendStartTLSMail(subject, content string, addressee []string) error {
	client, err := Dial(ms.AuthServer)
	if err != nil {
		return err
	}
	defer client.Close()
	if err = client.hello(); err != nil {
		return err
	}
	if ok, _ := client.Extension("STARTTLS"); ok {
		config := &tls.Config{ServerName: client.serverName, InsecureSkipVerify: true}
		if testHookStartTLS != nil {
			testHookStartTLS(config)
		}
		if err = client.StartTLS(config); err != nil {
			return fmt.Errorf("client startTLS error: %v", err)
		}
	}
	if ms.Auth != nil && client.ext != nil {
		if _, ok := client.ext["AUTH"]; !ok {
			return fmt.Errorf("smtp: server doesn't support AUTH")
		}
		if err = client.Auth(ms.Auth); err != nil {
			return fmt.Errorf("client auth error: %s ", err.Error())
		}
	}
	err = client.Mail(ms.SenderMail)
	if err != nil {
		return fmt.Errorf("client mail set from error: %v", err)
	}
	for _, to := range addressee {
		if err = client.Rcpt(to); err != nil {
			return fmt.Errorf("client rcpt %s error: %v", to, err)
		}
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("client data init error: %v", err)
	}
	_, err = w.Write(mailQQMessage(addressee, subject, content, ms.SenderName, ms.SenderMail))
	if err != nil {
		return fmt.Errorf("write message error: %v", err)
	}
	w.Close()
	err = client.Quit()
	if err != nil {
		return fmt.Errorf("client quit error: %v", err)
	}
	return err
}

func (ms *MailSender) sendTLSMail(subject, content string, addressee []string) error {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         ms.AuthServer,
	}
	conn, err := tls.Dial("tcp", ms.AuthServer, tlsConfig)
	if err != nil {
		return fmt.Errorf("tls dial error: %v", err)
	}
	client, newClientErr := NewClient(conn, ms.AuthServer)
	if newClientErr != nil {
		return fmt.Errorf("smtp new client error: %v", newClientErr)
	}
	defer client.Close()
	if b, _ := client.Extension("AUTH"); b {
		err = client.Auth(ms.Auth)
		if err != nil {
			return fmt.Errorf("client auth error: %v", err)
		}
	}
	err = client.Mail(ms.SenderMail)
	if err != nil {
		return fmt.Errorf("client mail set from error: %v", err)
	}
	for _, to := range addressee {
		if err = client.Rcpt(to); err != nil {
			return fmt.Errorf("client rcpt %s error: %v", to, err)
		}
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("client data init error: %v", err)
	}
	_, err = w.Write(mailQQMessage(addressee, subject, content, ms.SenderName, ms.SenderMail))
	if err != nil {
		return fmt.Errorf("write message error: %v", err)
	}
	w.Close()
	err = client.Quit()
	if err != nil {
		return fmt.Errorf("client quit error: %v", err)
	}
	return err
}

func mailQQMessage(addressee []string, subject, content, senderName, senderMail string) []byte {
	var buff bytes.Buffer
	buff.WriteString("To:")
	buff.WriteString(strings.Join(addressee, ","))
	buff.WriteString("\r\nFrom:")
	buff.WriteString(senderName + "<" + senderMail + ">")
	buff.WriteString("\r\nSubject:")
	buff.WriteString(subject)
	buff.WriteString("\r\nContent-Type:text/plain;charset=UTF-8\r\n\r\n")
	buff.WriteString(content)
	return buff.Bytes()
}

func VerifyMailAddress(mailString string) bool {
	reg := regexp.MustCompile(`\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*`)
	return reg.MatchString(mailString)
}
