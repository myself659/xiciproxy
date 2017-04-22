package xiciproxy

import (
	"database/sql"
	"fmt"
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

type ProxyPool struct {
	avaitems []string
	index    int
	allitems []string
	lock     sync.Mutex
}

func genProxyUrl(kind, ip, port string) string {
	if strings.Contains(kind, "https") || strings.Contains(kind, "HTTPS") {
		return "https://" + ip + ":" + port
	}

	return "http://" + ip + ":" + port
}

func New() *ProxyPool {
	var pool *ProxyPool
	user := "root"
	psword := "dbstar"
	ipport := "localhost:3306"
	dbname := "ippool"
	para := user + ":" + psword + "@tcp" + "(" + ipport + ")" + "/" + dbname + "?charset=utf8&parseTime=True"

	db, err := sql.Open("mysql", para)
	if err != nil {
		return pool
	}
	defer db.Close()

	rows, err := db.Query("SELECT IP, PORT,TYPE FROM xici_ip")
	if err != nil {
		return pool
	}

	pool = new(ProxyPool)
	pool.index = 0

	for rows.Next() {
		var ipaddr, port, kind string
		err := rows.Scan(ipaddr, port, kind)
		if err != nil {
			pool.allitems = append(pool.allitems, genProxyUrl(kind, ip, port))
		}
	}

	// 从数据库中获取所有数据，占用内存不会大大

	return pool
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
			time.Sleep(3600 * time.Second)

		}
		// 后续可以考虑Proxy复用，这样对GC友好
		request := gorequest.New().Proxy(proxyurl)
		resp, body, errs := request.Get(url).End()
		fmt.Println(errs)
		fmt.Println(resp)
		fmt.Println(body)
	}
}
