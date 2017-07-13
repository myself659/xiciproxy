package xiciproxy

import (
	"fmt"

	"github.com/PuerkitoBio/goquery"
	_ "github.com/go-sql-driver/mysql"
	"github.com/parnurzeal/gorequest"
	"math/rand"
	"net/http"
	"strings"
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

var uas = []string{"Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/57.0.2987.133 Safari/537.36",
	"Mozilla/5.0 (Windows; U; Windows NT 5.2) AppleWebKit/525.13 (KHTML, like Gecko) Version/3.1 Safari/525.13",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/56.0.2924.87 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; WOW64; rv:53.0) Gecko/20100101 Firefox/53.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/42.0.2311.135 Safari/537.36 Edge/12.246",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_2) AppleWebKit/601.3.9 (KHTML, like Gecko) Version/9.0.2 Safari/601.3.9",
	"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/47.0.2526.111 Safari/537.36",
	"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:15.0) Gecko/20100101 Firefox/15.0.1",
}

var ualen int = len(uas)

type proxyInfo struct {
	url string
	ua  string
}

type ProxyPool struct {
	urls     []string
	addurls  chan []string
	delurl   chan string
	getproxy chan *proxyInfo
	cur      *proxyInfo
	limit    int
}

func genProxyUrl(kind, ip, port string) string {
	if strings.Contains(kind, "https") || strings.Contains(kind, "HTTPS") {
		return "https://" + ip + ":" + port
	}

	return "http://" + ip + ":" + port
}

func NewProxyPool() *ProxyPool {
	pool := new(ProxyPool)

	pool.getproxy = make(chan *proxyInfo, 32)
	pool.urls = make([]string, 0)
	pool.addurls = make(chan []string)
	pool.delurl = make(chan string)
	pool.limit = 0
	pool.run()

	return pool
}

func (self *ProxyPool) fetch() {
	url := "http://www.xicidaili.com/nn"
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/57.0.2987.133 Safari/537.36"
	request := gorequest.New()
	resp, _, errs := request.Get(url).Set("User-Agent", ua).End()
	if errs != nil {
		fmt.Println(errs)
		return
	}

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		fmt.Println(err)
		return
	}

	/*
	 <table id="ip_list">
	*/
	tables := doc.Find("table")
	tableslen := tables.Length()
	//fmt.Println(tableslen)
	for i := 0; i < tableslen; i++ {
		table := tables.Eq(i)
		_, ok := table.Attr("id")
		//fmt.Println(idattr, ok)
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
				j := i - 1
				urls[j] = genProxyUrl(kind, ip, port)
			}
			fmt.Println("send urls")
			//fmt.Println(urls)
			self.addurls <- urls
			break

		}
	}
}

func (self *ProxyPool) run() {

	go func() {

		var temp proxyInfo
		for {

			if len(self.urls) < 50 {
				go self.fetch()
				<-time.After(3 * time.Second)
			}

			if len(self.urls) > 0 {
				urli := rand.Intn(len(self.urls))
				uai := rand.Intn(ualen)
				temp.url = self.urls[urli]
				temp.ua = uas[uai]
			}
			select {

			case urls := <-self.addurls:
				{

					fmt.Println("recv addurls:", len(urls))
					/*
						fmt.Println(urls)
					*/
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
							self.urls = self.urls[:len(self.urls)-1]
							//self.urls = append(self.urls[:i], self.urls[i+1:])
						}
					}

				}
			case self.getproxy <- &temp:
				{
					//do nothing
				}
			case <-time.After(100 * time.Second):
				{
					// time out
					fmt.Println("urls len:", len(self.urls))
				}
			}
		}
	}()

}

func (self *ProxyPool) Get(url string) (*http.Response, error) {
	var proxyurl string
	var uas string
	for {
		if self.limit <= 0 {
			pinfo := <-self.getproxy
			proxyurl = pinfo.url
			uas = pinfo.ua
			self.cur = pinfo
			self.limit = 60
		} else {
			proxyurl = self.cur.url
			uas = self.cur.ua
		}
		//fmt.Println(pinfo)

		// 后续可以考虑Proxy复用，这样对GC友好
		request := gorequest.New().Proxy(proxyurl).Set("User-Agent", uas).Timeout(60 * time.Second)
		resp, _, errs := request.Get(url).End()
		if errs == nil {
			self.limit--
			return resp, nil
		}
		fmt.Println(url, errs)
		// 有效提高速度,一定要将害群之马清除
		self.delurl <- proxyurl
		//fmt.Println(resp)
		self.limit = 0

	}
}

