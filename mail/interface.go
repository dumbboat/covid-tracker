package mail

import (
	netMail "net/mail"

	"github.com/dumbboat/covid-tracker/model"
)

type MailMessenger interface {
	Send(to, content string) error
}

type EXMailMessenger struct {
	mailBox model.Mailbox
}

func NewEXMailMessenger(mailBox model.Mailbox) EXMailMessenger {
	return EXMailMessenger{mailBox: mailBox}
}

func (m EXMailMessenger) Send(to, content string) error {
	from := netMail.Address{Name: "", Address: m.mailBox.User}
	sendto := netMail.Address{Name: "", Address: to}
	message := Setup(from.Address, sendto.Address, m.mailBox.Username)
	message += content
	client, err := Connect(m.mailBox.User, m.mailBox.Pwd)
	if err != nil {
		return err
	}
	return Send(from.Address, sendto.Address, client, []byte(message))
}
