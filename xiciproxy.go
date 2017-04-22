package xiciproxy

import (
	"database/sql"
	"fmt"
	"github.com/deckarep/golang-set"
	_ "github.com/go-sql-driver/mysql"
	"github.com/parnurzeal/gorequest"
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
	urls        []string
	urlindex    int
	uaindex     int
	tryindex    int
	trymax      int
	lock        sync.Mutex
	addurl      chan string
	delurl      chan string
	getproxy    chan proxyInfo
	fetchnotify chan bool
	num         int
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

func New(trymax int) *ProxyPool {
	pool := New(ProxyPool)

	pool.trymax = trymax
	pool.addurl = make(chan string)
	pool.delurl = make(chan string)
	pool.fetchnotify = make(chan bool)

	return pool
}

func (self *ProxyPool) dispatch()

func (self *ProxyPool) run() {

	go func() {
		for {
			select {
			case url := <-self.addurl:
				{
					append(self.urls, url)
				}
			case url := <-self.delurl:
				{
					// del invalid  url

				}
			case <-time.After(360 * time.Second):
				{
					if len(self.urls) < 100 {

					}
				}
			default:
				{
					if self.urlindex == len(self.urls) {
						self.urlindex = 0
					}
					for urli := self.urlindex; urli < len(self.urls); urli++ {
						if self.uaindex == ualen {
							self.uaindex = 0
						}

					}

				}
			}
		}
	}()

}

func (self *ProxyPool) getProxyUrl() string {
	self.lock.Lock()
	defer self.lock.Unlock()
	alllen := len(self.allitems)
	if alllen > 0 {
		tmp := self.allitems[alllen-1]
		self.allitems = self.allitems[0 : alllen-1]
		return tmp
	}

	avalen := len(self.avaitems)
	if avalen > 0 {
		if self.index == avalen {
			self.index = 0
		}
		tmp := self.avaitems[self.index]
		self.index++

		return tmp
	}

	return ""

}

func (self *ProxyPool) ProxyGet(url string) (http.Response, error) {

	for {
		proxyurl := self.getProxyUrl()
		if proxyurl == "" {
			// 通知获取goroutine去更新
			time.Sleep(360 * time.Second)

		}
		// 后续可以考虑Proxy复用，这样对GC友好
		request := gorequest.New().Proxy(proxyurl)
		resp, body, errs := request.Get(url).End()
		fmt.Println(errs)
		fmt.Println(resp)
		fmt.Println(body)
	}
}
