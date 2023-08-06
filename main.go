package main

import (
  "net/http"

  "github.com/gin-gonic/gin"
)

func main() {
  r := gin.Default()
  r.GET("/ping", func(c *gin.Context) {
    accessToken := c.Request.Header.Get("tnm-access-token")

    c.JSON(http.StatusOK, gin.H{
      "message": accessToken,
    })
  })

  // url redirect with custom header
  r.GET("/redirect", func(c *gin.Context) {
    accessToken := c.Request.Header.Get("tnm-access-token")

    c.Header("tnm-access-token", accessToken)
    c.Redirect(http.StatusMovedPermanently, "http://localhost:8080/ping")
  })

  r.Run()
}