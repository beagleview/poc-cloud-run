package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
)

var localCache = cache.New(5*time.Minute, 10*time.Minute)

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

	r.Any("/*any", func(c *gin.Context) {
		proxy(c)
	})

	r.Run()
}

func proxy(c *gin.Context) {
	authCodeKey := "authCode"
	accessTokenKey := "accessToken"
	expireInKey := "expireIn"

	if c.GetHeader("TMN-Auth-Code") != "" {
		localCache.Set(authCodeKey, c.GetHeader("TMN-Auth-Code"), cache.DefaultExpiration)
	}

	destURL := "xxx" + c.Request.URL.String()

	log.Println("destURL:", destURL)

	req, _ := http.NewRequest(c.Request.Method, destURL, c.Request.Body)

	// Copy headers from the original request to the proxy request
	for name, headers := range c.Request.Header {
		for _, h := range headers {
			req.Header.Add(name, h)
		}
	}

	client := &http.Client{}

	// Get access token
	if c.Request.URL.Path == "/paylater/dashboard/" {
		log.Println("request to: /paylater/dashboard/")

		accessTokenRep, found := localCache.Get(accessTokenKey)
		if !found || accessTokenRep == "" {
			authCode, found := localCache.Get(authCodeKey)
			if !found {
				c.String(http.StatusUnauthorized, "Unauthorized")
				return
			}

			reqData := RequestData{
				Code:         authCode.(string),
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

			reqAuth, err := http.NewRequest("POST", "xxx", strings.NewReader(form.Encode()))
			if err != nil {
				log.Fatal(err)
				return
			}

			reqAuth.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			respAuth, err := client.Do(reqAuth)
			if err != nil {
				log.Fatal(err)
				return
			}
			defer respAuth.Body.Close()

			bodyAuth, err := io.ReadAll(respAuth.Body)
			if err != nil {
				log.Fatal(err)
				return
			}

			log.Println("response Body:", string(bodyAuth))

			var resAuthData ResponseData
			json.Unmarshal(bodyAuth, &resAuthData)
			log.Println("access token:", resAuthData.AccessToken)

			if resAuthData.AccessToken != "" {
				log.Println("set access token to cache")
				localCache.Set(accessTokenKey, resAuthData.AccessToken, cache.DefaultExpiration)
				localCache.Set(expireInKey, resAuthData.ExpiresIn, cache.DefaultExpiration)
			}
		}
	}
	//-------

	accessTokenRep, found := localCache.Get(accessTokenKey)
	if found {
		log.Println("set access token to request header: " + accessTokenRep.(string))
		req.Header.Set("tmn-access-token", accessTokenRep.(string))
	}

	expireInRep, found := localCache.Get(expireInKey)
	if found {
		req.Header.Set("tmn-expires-in", expireInRep.(string))
	}

	resp, err := client.Do(req)
	if err != nil {
		c.String(http.StatusBadGateway, err.Error())
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// Copy headers from the proxy response to the client response
	for name, headers := range resp.Header {
		for _, h := range headers {
			c.Writer.Header().Add(name, h)
		}
	}

	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
}

func authUserHandler(c *gin.Context) {
	authCode := c.GetHeader("TMN-Auth-Code")

	reqData := RequestData{
		Code:         authCode,
		GrantType:    "authorization_code",
		ClientId:     "gSJqj6PRnUhTsyppVZO0WlY7Lf0oUK",
		ClientSecret: "Ua4yf7w2li2LqPcFHzSypWOxXm0kfL",
		RedirectUri:  "https://asia-southeast1-poc-gcloud-function.cloudfunctions.net/auth-user",
	}

	form := url.Values{}
	form.Add("code", reqData.Code)
	form.Add("grant_type", reqData.GrantType)
	form.Add("client_id", reqData.ClientId)
	form.Add("client_secret", reqData.ClientSecret)
	form.Add("redirect_uri", reqData.RedirectUri)

	req, err := http.NewRequest("POST", "https://apis.tmn-dev.com/oauth2/v1/token", strings.NewReader(form.Encode()))
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
	log.Println("access token:", resData.AccessToken)

	feReq, err := http.NewRequest("GET", "https://lending-dashboard-qa.public-a-cloud1n.ascendnano.io/paylater/dashboard/", nil)
	if err != nil {
		log.Fatal(err)
		return
	}

	feReq.Header.Set("tmn-access-token", resData.AccessToken)
	feReq.Header.Set("tmn-expires-in", resData.ExpiresIn)

	feResp, err := client.Do(feReq)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer feResp.Body.Close()

	for name, values := range feResp.Header {
		for _, value := range values {
			c.Header(name, value)
		}
	}

	//c.Data(feResp.StatusCode, feResp.Header.Get("Content-Type"), feBody)
	io.Copy(c.Writer, feResp.Body)
}
