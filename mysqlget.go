package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/parnurzeal/gorequest"
	"sync"
)

type ProxyItem struct {
	ipaddr string
	port   string
}

type ProxyPool struct {
	avaitems []string
	index    int
	allitems []ProxyItem
	lock     sync.Mutex
}

func proxyget(url string, proxyurl string) {
	request := gorequest.New().Proxy(proxyurl)
	resp, body, errs := request.Get(url).End()
	fmt.Println(errs)
	fmt.Println(resp)
	fmt.Println(body)
}
func main() {
	var pool *ProxyPool
	user := "root"
	psword := "dbstar"
	ipport := "localhost:3306"
	dbname := "ippool"
	para := user + ":" + psword + "@tcp" + "(" + ipport + ")" + "/" + dbname + "?charset=utf8&parseTime=True"

	db, err := sql.Open("mysql", para)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT IP, PORT FROM xici_ip")
	if err != nil {
		fmt.Println(err)
		return
	}

	pool = new(ProxyPool)
	pool.index = 0

	for rows.Next() {
		var temp ProxyItem
		err := rows.Scan(&temp.ipaddr, &temp.port)
		if err == nil {
			proxyurl := "http://" + temp.ipaddr + ":" + temp.port
			proxyget("https://www.baidu.com", proxyurl)
			fmt.Println(temp)
			pool.allitems = append(pool.allitems, temp)

		}
	}

	// 从数据库中获取所有数据，占用内存不会大大

	return
}
