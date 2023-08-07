package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

type RequestData struct {
	Code         string `url:"code"`
	GrantType    string `url:"grant_type"`
	ClientId     string `url:"client_id"`
	ClientSecret string `url:"client_secret"`
	RedirectUri  string `url:"redirect_uri"`
}

type ResponseData struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
	ExpiresIn    string `json:"expires_in"`
}

func main() {
	r := gin.Default()
	r.GET("/auth-user", authUserHandler)

	// url redirect with custom header
	r.GET("/redirect", func(c *gin.Context) {
		accessToken := c.GetHeader("tmn-access-token")
		c.JSON(http.StatusOK, gin.H{
			"access_token": accessToken,
		})
	})

	r.GET("/test", func(c *gin.Context) {
		c.Header("tmn-access-token", "123456789000fghgjkdhk")
		c.SetCookie("myCookie", "myValue", 3600, "/", "", false, true)
		c.Redirect(http.StatusMovedPermanently, "http://localhost:8080/redirect")
	})

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	r.Run()
}

func authUserHandler(c *gin.Context) {
	authCode := c.GetHeader("TMN-Auth-Code")

	reqData := RequestData{
		Code:         authCode,
		GrantType:    "authorization_code",
		ClientId:     "xxx",
		ClientSecret: "xxx",
		RedirectUri:  "xxx",
	}

	form := url.Values{}
	form.Add("code", reqData.Code)
	form.Add("grant_type", reqData.GrantType)
	form.Add("client_id", reqData.ClientId)
	form.Add("client_secret", reqData.ClientSecret)
	form.Add("redirect_uri", reqData.RedirectUri)

	req, err := http.NewRequest("POST", "xxx", strings.NewReader(form.Encode()))
	if err != nil {
		log.Fatal(err)
		return
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	log.Println("response Body:", string(body))

	var resData ResponseData
	json.Unmarshal(body, &resData)

	c.Header("tmn-access-token", resData.AccessToken)
	c.Header("tmn-expires-in", resData.ExpiresIn)

	log.Println("access token:", resData.AccessToken)

	//c.Redirect(http.StatusMovedPermanently, "https://lending-dashboard-qa.public-a-cloud1n.ascendnano.io/paylater/dashboard/")
	c.Redirect(http.StatusMovedPermanently, "http://localhost:8080/redirect")
}
