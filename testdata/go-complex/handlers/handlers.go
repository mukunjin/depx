package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

// 别名导入
var rootCmd = cobra.Command{
	Use: "test",
}

func PingHandler(c *gin.Context) {
	c.JSON(200, gin.H{"message": "pong"})
}
