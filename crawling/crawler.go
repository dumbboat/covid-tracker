package crawling

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"time"

	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
)

const (
	_url = "http://wsjkw.sh.gov.cn/yqtb/index.html" //上海市卫健委疫情发布
	_ua  = "Mozilla/5.0 (Windows NT 6.3; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.103 Safari/537.36"
)

const (
	ReportSuffix = "sh-covid19-report.html"
)

func CrawlShanghaiCovid19Report(filename string) error {
	html, err := scrapeNewArticleHtml(_url)
	if err != nil {
		return fmt.Errorf("[ERROR] Failed to scrapeNewArticleHtml:%s", err.Error())
	}
	return ioutil.WriteFile(filename, []byte(html), 0644)
}

func scrapeNewArticleHtml(url string) (html string, err error) {
	options := []chromedp.ExecAllocatorOption{
		chromedp.Flag("headless", false), // 是否打开浏览器调试
		chromedp.UserAgent(_ua),          // 设置User-Agent
	}
	options = append(chromedp.DefaultExecAllocatorOptions[:], options...)

	var allocCtx context.Context
	var cancel context.CancelFunc
	if checkChromePort() {
		allocCtx, cancel = chromedp.NewRemoteAllocator(context.Background(), "ws://127.0.0.1:9222/")
	} else {
		allocCtx, cancel = chromedp.NewExecAllocator(context.Background(), options...)
	}
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()
	// set timeout
	ctx, cancel = context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	// listening target ID of the second tab
	ch := make(chan target.ID, 1)
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		if ev, ok := ev.(*target.EventTargetCreated); ok &&
			// if OpenerID == "", this is the first tab.
			
			ev.TargetInfo.OpenerID != "" {
			ch <- ev.TargetInfo.TargetID
		}
	})
	log.Println(url)
	var body string
	if err = chromedp.Run(ctx,
		chromedp.Tasks{
			chromedp.Navigate(url),
			chromedp.WaitVisible("body"),
			chromedp.Sleep(5 * time.Second),
			chromedp.Click(`.uli16 > li:nth-child(1)`, chromedp.ByQuery),
			chromedp.Sleep(5 * time.Second),
			chromedp.OuterHTML("html", &body, chromedp.ByQuery),
		},
	); err != nil {
		log.Printf("[ERROR] chromedp failed: %s", err.Error())
		return
	}

	newCtx, cancel := chromedp.NewContext(ctx, chromedp.WithTargetID(<-ch))
	defer cancel()
	if err = chromedp.Run(
		newCtx,
		chromedp.Sleep(2*time.Second),
		chromedp.OuterHTML("#js_content", &html, chromedp.ByID),
	); err != nil {
		log.Printf("[ERROR] chromedp failed on the second tab: %s", err.Error())
		return
	}
	return html, nil
}

//检查是否有9222端口，来判断是否运行在linux上
func checkChromePort() bool {
	addr := net.JoinHostPort("", "9222")
	conn, err := net.DialTimeout("tcp", addr, 1*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}


