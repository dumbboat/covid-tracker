package store

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

const store = "Addr2EmailStore.store"

var addr2EmailStore map[string] /*addr*/ map[string] /*email addrress*/ struct{}

func init() {
	addr2EmailStore = make(map[string]map[string]struct{})
	bs, err := os.ReadFile(store)
	if err != nil {
		log.Printf("[ERROR] Failed to read store file:%s", err.Error())
	} else {
		err = json.Unmarshal(bs, &addr2EmailStore)
		log.Printf("loading Addr2EmailStore,err:%v", err)
	}
	go periodicalPersisting()
}

func GetStore() map[string]map[string]struct{} {
	return addr2EmailStore
}

func Append(addr string, email string) {
	if emails, exists := addr2EmailStore[addr]; exists {
		emails[email] = struct{}{}
	} else {
		addr2EmailStore[addr] = make(map[string]struct{})
		addr2EmailStore[addr][email] = struct{}{}
	}
}

func Delete(addr string, email string) {
	if emails, exists := addr2EmailStore[addr]; exists {
		delete(emails, email)
	}
}

func Persist() error {
	bs, err := json.Marshal(&addr2EmailStore)
	if err != nil {
		return fmt.Errorf("[ERROR]Failed to marshal Addr2EmailStore:%s", err.Error())
	}
	return os.WriteFile(store, bs, 0644)
}

func periodicalPersisting() {
	ticker := time.NewTicker(time.Minute * 10)
	for {
		<-ticker.C
		Persist()
	}
}
