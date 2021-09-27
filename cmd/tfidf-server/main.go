package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/Sudalight/tools/pkg/tfidf"
	"github.com/gin-gonic/gin"
)

var (
	storeFilename = flag.String("fn", "tfidf.json", "filename of tfidf persistent data")
	port          = flag.String("p", "12345", "service port")
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

	server, err := tfidf.NewServer(*storeFilename)
	if err != nil {
		panic(err)
	}
	router.POST("/upsert_docs", server.UpsertDocs)
	router.POST("/get_doc_vector", server.GetDocVector)

	sigterm := make(chan os.Signal, 1)
	go func() {
		for {
			timer := time.Tick(time.Minute)
			select {
			case <-timer:
				err = server.Save(*storeFilename)
				if err != nil {
					log.Println(err)
				} else {
					log.Println("auto saved successfully!")
					runtime.GC()
				}
			case <-sigterm:
				err = server.Save(*storeFilename)
				if err != nil {
					log.Println(err)
				} else {
					log.Println("auto saved before exit")
				}
				os.Exit(0)
			}
		}
	}()

	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	log.Println("ready perfectly!")
	panic(router.Run(":" + *port))
}
