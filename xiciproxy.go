package xiciproxy

import (
	"fmt"

	"github.com/PuerkitoBio/goquery"
	_ "github.com/go-sql-driver/mysql"
	"github.com/parnurzeal/gorequest"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

//http request

/*
User-Agent:Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/57.0.2987.133 Safari/537.36

UserAgent = [
'Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.0)',
'Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 5.2)',
'Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1)',
'Mozilla/5.0 (Windows; U; Windows NT 5.2) Gecko/2008070208 Firefox/3.0.1',
'Mozilla/5.0 (Windows; U; Windows NT 5.1) Gecko/20070803 Firefox/1.5.0.12',
'Mozilla/5.0 (Macintosh; PPC Mac OS X; U; en) Opera 8.0',
'Opera/8.0 (Macintosh; PPC Mac OS X; U; en)',
'Opera/9.27 (Windows NT 5.2; U; zh-cn)',
'Mozilla/5.0 (Windows; U; Windows NT 5.2) AppleWebKit/525.13 (KHTML, like Gecko) Chrome/0.2.149.27 Safari/525.13',
'Mozilla/5.0 (Windows; U; Windows NT 5.1; en-US; rv:1.8.1.12) Gecko/20080219 Firefox/2.0.0.12 Navigator/9.0.0.6',
'Mozilla/5.0 (iPhone; U; CPU like Mac OS X) AppleWebKit/420.1 (KHTML, like Gecko) Version/3.0 Mobile/4A93 Safari/419.3',
'Mozilla/5.0 (Windows; U; Windows NT 5.2) AppleWebKit/525.13 (KHTML, like Gecko) Version/3.1 Safari/525.13'
]


*/
/*
透明代理
代理的策略 边测边验证的方式先验证再循环的循环模式
*/

var ua = []string{"Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/57.0.2987.133 Safari/537.36",
	"Mozilla/5.0 (Windows; U; Windows NT 5.2) AppleWebKit/525.13 (KHTML, like Gecko) Version/3.1 Safari/525.13",
	"Opera/9.27 (Windows NT 5.2; U; zh-cn)",
}

var ualen int = len(ua)

type ProxyPool struct {
	urls []string

	lock     sync.Mutex
	addurls  chan []string
	delurl   chan string
	getproxy chan *proxyInfo
	cycles   int
}

type proxyInfo struct {
	url string
	ua  string
}

func genProxyUrl(kind, ip, port string) string {
	if strings.Contains(kind, "https") || strings.Contains(kind, "HTTPS") {
		return "https://" + ip + ":" + port
	}

	return "http://" + ip + ":" + port
}

func NewProxyPool() *ProxyPool {
	pool := new(ProxyPool)

	pool.addurls = make(chan []string)
	pool.delurl = make(chan string)
	pool.urls = make([]string, 4096)
	pool.getproxy = make(chan *proxyInfo, 32)
	pool.run()

	return pool
}

func (self *ProxyPool) fetch() {
	url := "http://www.xicidaili.com/nn"
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/57.0.2987.133 Safari/537.36"
	request := gorequest.New()
	resp, _, errs := request.Get(url).Set("User-Agent", ua).End()
	fmt.Println(resp)

	fmt.Println(errs)

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(resp.Body)

	/*
	 <table id="ip_list">
	*/
	tables := doc.Find("table")
	tableslen := tables.Length()
	fmt.Println(tableslen)
	for i := 0; i < tableslen; i++ {
		table := tables.Eq(i)
		idattr, ok := table.Attr("id")
		fmt.Println(idattr, ok)
		if ok == true {
			trs := table.Find("tr")
			trslen := trs.Length()
			urls := make([]string, trslen)
			for i := 1; i < trslen; i++ {
				tr := trs.Eq(i)
				tds := tr.Find("td")
				kind := tds.Eq(5).Text()
				ip := tds.Eq(1).Text()
				port := tds.Eq(2).Text()
				fmt.Println(kind)
				fmt.Println(ip)
				fmt.Println(port)
				urls = append(urls, genProxyUrl(kind, ip, port))
			}

			self.addurls <- urls
		}
	}
}

func (self *ProxyPool) run() {

	go func() {
		for {
			select {
			case urls := <-self.addurls:
				{
					for _, url := range urls {
						self.urls = append(self.urls, url)
					}
				}
			case url := <-self.delurl:
				{
					// del invalid  url
					for i := 0; i < len(self.urls); i++ {
						if self.urls[i] == url {
							copy(self.urls[i:], self.urls[i+1:])
						}
					}

				}
			case <-time.After(60 * time.Second):
				{
					self.cycles++
					if len(self.urls) < 100 {
						go self.fetch()
					} else if self.cycles%60 == 0 {
						go self.fetch()
					}
				}

			default:
				{
					urli := rand.Intn(len(self.urls))
					uai := rand.Intn(ualen)
					pinfo := &proxyInfo{url: self.urls[urli], ua: ua[uai]}
					self.getproxy <- pinfo
				}
			}
		}
	}()

}

func (self *ProxyPool) ProxyGet(url string) (http.Response, error) {

	for {
		pinfo := <-self.getproxy
		proxyurl := pinfo.url
		uas := pinfo.ua
		// 后续可以考虑Proxy复用，这样对GC友好
		request := gorequest.New().Proxy(proxyurl).Set("User-Agent", uas)
		resp, _, errs := request.Get(url).End()
		if errs == nil {
			return http.Response(*resp), nil
		}
		fmt.Println(errs)
		fmt.Println(resp)

	}
}
