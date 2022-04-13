package mail

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
)

func Setup(from, to, sender string) string {
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = fmt.Sprintf("您收到了一条来自 %s 的消息", "上海市新冠疫情订阅")
	body := ""
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body
	return message
}

func Send(from, to string, client *smtp.Client, mailBody []byte) (err error) {
	if err = client.Mail(from); err != nil {
		return err
	}
	if err = client.Rcpt(to); err != nil {
		return err
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	_, err = writer.Write(mailBody)
	err = writer.Close()
	if err != nil {
		return err
	}
	err = client.Quit()
	if err != nil {
		return err
	}
	return nil
}

func Connect(username, password string) (*smtp.Client, error) {
	servername := "smtp.exmail.qq.com:465"
	host, _, _ := net.SplitHostPort(servername)
	auth := smtp.PlainAuth("", username, password, host)
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}
	conn, err := tls.Dial("tcp", servername, tlsconfig)
	if err != nil {
		return nil, err
	}
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return nil, err
	}
	if err = client.Auth(auth); err != nil {
		return nil, err
	}
	return client, nil
}
