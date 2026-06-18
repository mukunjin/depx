package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"golang.org/x/text/language"
)

func main() {
	r := gin.Default()
	
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})
	
	// 使用 redis
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	_ = rdb
	
	// 使用 text
	tag := language.English
	fmt.Println(tag)
	
	r.Run(":8080")
}
