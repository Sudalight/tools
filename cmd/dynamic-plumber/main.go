package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"

	dynamicplumber "github.com/Sudalight/tools/pkg/dynamic-plumber"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

var (
	port = flag.Int("p", 12345, "service port")
)

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Llongfile)

	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()
	router.GET("/check", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "ready perfectly",
		})
	})

	router.GET("/run", func(ctx *gin.Context) {
		err := dynamicplumber.LoadPlugin(ctx.Query("name"))
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, err.Error())
			return
		}
	})

	pprof.Register(router)
	log.Println("ready perfectly!")
	panic(router.Run(":" + strconv.Itoa(*port)))
}
