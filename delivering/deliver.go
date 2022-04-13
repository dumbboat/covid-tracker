package delivering

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/dumbboat/covid-tracker/mail"
	"github.com/dumbboat/covid-tracker/model"
	"github.com/dumbboat/covid-tracker/store"
)

func Deliver(mailboxConfigPath string, content []byte) error {
	mailBox := mail.NewEXMailMessenger(model.GetMailboxFromConf(mailboxConfigPath))
	brief, addrs, err := ParseData(content, "p")
	if err != nil {
		return fmt.Errorf("[ERROR] Failed to GetAddressData: %s", err.Error())
	}
	s := store.GetStore()
	for k, emails := range s {
		var possibleAddrs []string
		for i := range addrs {
			if strings.Contains(addrs[i], k) {
				possibleAddrs = append(possibleAddrs, addrs[i])
			}
		}

		var result string
		if len(possibleAddrs) == 0 {
			result = fmt.Sprintf("您所在的地址 %s 未发现有新增阳性感染者", k)
		} else {
			result = fmt.Sprintf("您所在的地址: %s\n下面的地址有新增阳性感染者:\n%s", k, strings.Join(possibleAddrs, "\n"))
		}

		for mailBoxAddr, _ := range emails {
			emailContent := fmt.Sprintf(
				`
				%s
				
				上海市各区感染情况:
	
				%s
				
				
				<a href="http://dboat.cn/unregister?email=%s&&addr=%s">点击取消订阅</a>
				`, result, brief, mailBoxAddr, k)
			if err = mailBox.Send(mailBoxAddr, emailContent); err != nil {
				log.Printf("[ERROR] Sending email %s to %s(addr:%s) failed:%s", emailContent, mailBoxAddr, k, err.Error())
			}

		}
	}
	return nil
}

func ParseData(htmlContent []byte, selector string) (string, []string, error) {
	var addrs []string
	var builder strings.Builder
	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(htmlContent))
	if err != nil {
		log.Print(err.Error())
		return "", nil, err
	}

	livesAtSuffix := "分别居住于："
	dom.Find(selector).Each(func(i int, selection *goquery.Selection) {
		text := strings.TrimSpace(selection.Text())

		if strings.HasSuffix(text, livesAtSuffix) {
			trimed := strings.Replace(text, livesAtSuffix, "\n", -1)
			builder.WriteString(strings.Replace(trimed, "，", ",", -1) + "\n")
		} else {
			text = strings.TrimSuffix(text, "，")
			text = strings.TrimSuffix(text, "。")
			addrs = append(addrs, text)
		}
	})
	return builder.String(), addrs, nil
}
