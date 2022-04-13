package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dumbboat/covid-tracker/delivering"
	"github.com/dumbboat/covid-tracker/store"
	"github.com/dumbboat/covid-tracker/wechat"
	"github.com/thedevsaddam/renderer"
)

var rnd *renderer.Render
var configFile *string

func init() {
	opts := renderer.Options{
		ParseGlobPattern: "./tpl/*.html",
	}

	rnd = renderer.New(opts)
}

func main() {
	configFile = flag.String("c", "./exmail.conf", "it's the path to the config file that covid-tracker uses")
	flag.Parse()

	// setup signal catching
	sigs := make(chan os.Signal, 1)

	// catch all signals since not explicitly listing
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	//signal.Notify(sigs,syscall.SIGQUIT)

	// method invoked upon seeing signal
	go func() {
		s := <-sigs
		log.Printf("Received signal: %s, Exiting...", s)
		store.Persist() // persisting storage
		os.Exit(0)
	}()

	loc, _ := time.LoadLocation("Asia/Shanghai")
	ticker := time.NewTicker(time.Minute * 1)
	go func() {
		var lastDeliveredDate string
		for {
			<-ticker.C

			now := time.Now()
			date := now.Format("2006-01-02")
			yesterday := now.AddDate(0, 0, -1).Format("2006年1月02日")
			if date == lastDeliveredDate {
				continue
			}

			head := time.Date(now.Year(), now.Month(), now.Day(), 10, 0, 0, 0, loc)
			tail := time.Date(now.Year(), now.Month(), now.Day(), 13, 0, 0, 0, loc)
			if now.Before(head) || now.After(tail) {
				continue
			}

			filename := fmt.Sprintf("./%s-%s", date, wechat.ReportSuffix)
			err := wechat.CrawlShanghaiCovid19Report(filename)
			if err != nil {
				log.Printf("crawled daily covid19 report of shanghai,err: %s\n", err.Error())
			} else {
				bs, err := os.ReadFile(filename)
				if err != nil {
					log.Printf("[ERROR] Failed to read %s: %s", filename, err.Error())
				}
				match := bytes.Contains(bs, []byte(yesterday))
				log.Printf("日期(%s)匹配:%v", yesterday, match)
				if match {
					err = delivering.Deliver(*configFile, bs)
					if err == nil {
						lastDeliveredDate = date
					}
				}
			}
		}
	}()
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("assets"))
	mux.Handle("/", fs)
	mux.HandleFunc("/about", about)
	mux.HandleFunc("/register", Register)
	mux.HandleFunc("/unregister", UnRegister)
	port := ":80"
	log.Println("Listening on port ", port)
	http.ListenAndServe(port, mux)

}

func about(w http.ResponseWriter, r *http.Request) {
	rnd.HTML(w, http.StatusOK, "about", nil)
}

func Register(w http.ResponseWriter, r *http.Request) {
	result := struct {
		Result string
	}{""}
	addr := r.FormValue("addr")
	email := r.FormValue("email")
	if addr != "" && email != "" {
		store.Append(addr, email)
		result.Result = "订阅成功"
	}
	rnd.HTML(w, http.StatusOK, "home", result)
}

func UnRegister(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()
	email := values.Get("email")
	addr := values.Get("addr")
	store.Delete(addr, email)
	rnd.HTMLString(w, http.StatusOK, "<div>取消订阅成功,如果您错误操作请<a href='http://dboat.cn'>重新登记</a></div>")
}
