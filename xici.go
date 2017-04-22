package main

import (
	_ "crypto/tls"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/parnurzeal/gorequest"
	_ "net/http"
)

/*
User-Agent:Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/57.0.2987.133 Safari/537.36
*/
func fetch(url string) {
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/57.0.2987.133 Safari/537.36"
	request := gorequest.New()
	resp, _, errs := request.Get(url).Set("User-Agent", ua).End()
	fmt.Println(resp)
	//fmt.Println(body)
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
			for i := 1; i < trslen; i++ {
				tr := trs.Eq(i)
				tds := tr.Find("td")
				fmt.Println(tds.Eq(1).Text())
				fmt.Println(tds.Eq(2).Text())
				fmt.Println(tds.Eq(5).Text())
			}
		}
	}
}

func main() {
	url := "http://www.xicidaili.com/nn"
	fetch(url)
}
