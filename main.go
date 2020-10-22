package main

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/robertkrimen/otto"
)

type Login struct {
	User     string `json:"user"`
	Password string `json:"pass"`
}

func main() {
	r := gin.Default()
	
	r.POST("/sign", func(c *gin.Context) {
		var json Login
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		fmt.Println(json)
		c.JSON(http.StatusOK, gin.H{"status": sign(json)})
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func sign(json Login) interface{} {

	client := http.Client{}
	baseURL, _ := url.Parse("https://portal.ncu.edu.tw/")
	jar, _ := cookiejar.New(nil)
	client.Jar = jar
	form := make(url.Values)

	client.Get(baseURL.String())
	form.Add("_csrf", client.Jar.Cookies(baseURL)[0].Value)
	form.Add("username", json.User)
	form.Add("password", json.Password)
	form.Add("language", "CHINESE")

	client.PostForm("https://portal.ncu.edu.tw/login", form) // login

	res, _ := client.Get(baseURL.String())
	doc, _ := goquery.NewDocumentFromResponse(res)

	// Get Token After Login
	s := doc.Find("body > footer:nth-child(8) > script")
	vm := otto.New()
	vm.Run(s.Text())
	val, _ := vm.Get("token")

	// Login to HumanSys
	LoginHumanSys, _ := url.Parse("https://portal.ncu.edu.tw/system/142")
	q, _ := url.ParseQuery(LoginHumanSys.RawQuery)
	q.Set("token", val.String())
	LoginHumanSys.RawQuery = q.Encode()
	client.Get(LoginHumanSys.String())

	//Get WorkList
	Worklist := "https://cis.ncu.edu.tw/HumanSys/student/stdSignIn/"
	res, _ = client.Get(Worklist)
	doc, _ = goquery.NewDocumentFromResponse(res)

	//Go First Work
	s = doc.Find("#table1 > tbody > tr:nth-child(2) > td:nth-child(6) > a:nth-child(1)")
	link, _ := s.Attr("href")
	res, _ = client.Get("https://cis.ncu.edu.tw/HumanSys/student/" + link)
	signpage, _ := goquery.NewDocumentFromResponse(res)
	idno, _ := signpage.Find("#idNo").Attr("value")
	ParttimeUsuallyId, _ := signpage.Find("#ParttimeUsuallyId").Attr("value")
	token, _ := signpage.Find("body > div.container-fluid > div > input[type=hidden]").Attr("value")
	signform := make(url.Values) // build sign form
	signform.Set("functionName", "doSign")
	signform.Add("idNo", idno)
	signform.Add("ParttimeUsuallyId", ParttimeUsuallyId)
	signform.Add("AttendWork", "機房值班")
	signform.Add("_token", token)
	// res, _ = client.PostForm("https://cis.ncu.edu.tw/HumanSys/student/stdSignIn_detail", signform)
	return signform
}
