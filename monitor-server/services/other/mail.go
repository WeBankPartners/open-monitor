package other

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"go.uber.org/zap"
	"net"
	"net/smtp"
	"strings"
	"time"
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
	_ = SendSmtpMailResult(smo)
}

// SendSmtpMailResult is the error-returning path used by durable notification
// workers. The legacy SendSmtpMail function remains fire-and-forget for old
// callers.
func SendSmtpMailResult(smo m.SendAlertObj) error {
	if !mailEnable {
		return nil
	}
	if m.Config().Alert.Mail.Tls {
		return sendSMTPMailTLSResult(smo)
	}
	err := sendSMTPMailResult(smo, false)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Send mail error", zap.Error(err))
	}
	return err
}

func sendSMTPMailTLS(smo m.SendAlertObj) {
	_ = sendSMTPMailTLSResult(smo)
}

func sendSMTPMailTLSResult(smo m.SendAlertObj) error {
	return sendSMTPMailResult(smo, true)
}

// Bound connection establishment, TLS negotiation and all SMTP commands. A
// stalled relay must not hold the durable notification worker indefinitely.
func newSMTPClient(address string, implicitTLS bool, timeout time.Duration) (*smtp.Client, error) {
	deadline := time.Now().Add(timeout)
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return nil, err
	}
	if err = conn.SetDeadline(deadline); err != nil {
		conn.Close()
		return nil, err
	}
	host, _, _ := net.SplitHostPort(address)
	if implicitTLS {
		conn = tls.Client(conn, &tls.Config{InsecureSkipVerify: true, ServerName: host})
	}
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		conn.Close()
	}
	return client, err
}

func sendSMTPMailResult(smo m.SendAlertObj, implicitTLS bool) error {
	tlsConfig := &tls.Config{
		ServerName: smtpServer,
	}
	port := "25"
	if implicitTLS {
		port = "465"
	}
	client, err := newSMTPClient(net.JoinHostPort(smtpServer, port), implicitTLS, 10*time.Second)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Smtp new client error", zap.Error(err))
		return err
	}
	defer client.Close()
	if !implicitTLS {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err = client.StartTLS(tlsConfig); err != nil {
				return err
			}
		}
		if ok, _ := client.Extension("AUTH"); smtpAuth != nil && !ok {
			return errors.New("SMTP server does not support AUTH")
		}
	}
	if b, _ := client.Extension("AUTH"); b && smtpAuth != nil {
		err = client.Auth(smtpAuth)
		if err != nil {
			log.Error(nil, log.LOGGER_APP, "Client auth error", zap.Error(err))
			return err
		}
	}
	err = client.Mail(sendFrom)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Client mail set from error", zap.Error(err))
		return err
	}
	for _, to := range smo.Accept {
		if err = client.Rcpt(to); err != nil {
			log.Error(nil, log.LOGGER_APP, fmt.Sprintf("Client rcpt %s error", to), zap.Error(err))
			return err
		}
	}
	w, err := client.Data()
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Client data init error", zap.Error(err))
		return err
	}
	_, err = w.Write(mailQQMessage(smo.Accept, smo.Subject, smo.Content))
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Write message error", zap.Error(err))
		return err
	}
	if err = w.Close(); err != nil {
		log.Error(nil, log.LOGGER_APP, "Close message error", zap.Error(err))
		return err
	}
	err = client.Quit()
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Client quit error", zap.Error(err))
		return err
	}
	return nil
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
