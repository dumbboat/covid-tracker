package model

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type Mailbox struct {
	Host               string
	TLS                bool
	InsecureSkipVerify bool
	User               string
	Pwd                string
	Folder             string
	// Read only mode, false (original logic) if not initialized
	ReadOnly bool
	Username string
}

func GetMailboxFromConf(filepath string) (mailbox Mailbox) {
	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Panic(err)
	}

	err = json.Unmarshal(content, &mailbox)
	if err != nil {
		log.Panic(err)
	}
	return
}
