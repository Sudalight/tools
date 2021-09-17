package tfidf

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type Server struct {
	tfidf *TFIDF
}

func NewServer(filename string) (*Server, error) {
	s := &Server{
		tfidf: NewTFIDF(),
	}
	_, err := os.Stat(filename)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	} else if os.IsNotExist(err) {
		err = ioutil.WriteFile(filename, []byte("{}"), 0777)
		if err != nil {
			return nil, err
		}
	}

	err = s.tfidf.LoadFrom(filename)
	return s, err
}

func (s *Server) Save(filename string) error {
	return s.tfidf.Save(filename)
}

func (s *Server) UpsertDocs(ctx *gin.Context) {
	req := []Doc{}
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		log.Println(err)
		ctx.JSON(http.StatusBadRequest, fmt.Sprintf("invalid parameters, %s", err.Error()))
		return
	}

	s.tfidf.UpsertDocs(req)
	ctx.JSON(http.StatusOK, "ok")
}

func (s *Server) GetDocVector(ctx *gin.Context) {
	req := Doc{}
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		log.Println(err)
		ctx.JSON(http.StatusBadRequest, fmt.Sprintf("invalid parameters, %s", err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, s.tfidf.GetDocVector(req))
}
