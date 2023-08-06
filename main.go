package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RequestData struct {
	Code         string `json:"code"`
	GrantType    string `json:"grant_type"`
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectUri  string `json:"redirect_uri"`
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
		accessToken := c.Request.Header.Get("tnm-access-token")

		c.JSON(http.StatusOK, gin.H{
			"access_token": accessToken,
		})
	})

	r.Run()
}

func authUserHandler(c *gin.Context) {
	authCode := c.Request.Header.Get("TMN-Auth-Code")

	reqData := RequestData{
		Code:         authCode,
		GrantType:    "authorization_code",
		ClientId:     "client_id",
		ClientSecret: "client_secret",
		RedirectUri:  "http://localhost:8080/redirect",
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		log.Fatal(err)
		return
	}

	req, err := http.NewRequest("POST", "https://apis.tmn-dev.com/oauth2/v1/token", bytes.NewBuffer(jsonData))
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

	body, _ := ioutil.ReadAll(resp.Body)

	log.Println("response Body:", string(body))

	var resData ResponseData
	json.Unmarshal(body, &resData)

	c.Header("tmn-access-token", resData.AccessToken)
	c.Header("tmn-expires-in", resData.ExpiresIn)

	c.Redirect(http.StatusMovedPermanently, "http://localhost:8080/redirect")
}
