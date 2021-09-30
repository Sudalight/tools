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

func NewServer(pdFilename, fdFilename string) (*Server, error) {
	s := &Server{
		tfidf: NewTFIDF(),
	}
	_, err := os.Stat(pdFilename)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	} else if os.IsNotExist(err) {
		err = ioutil.WriteFile(pdFilename, []byte("{}"), 0777)
		if err != nil {
			return nil, err
		}
	}

	_, err = os.Stat(fdFilename)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	} else if os.IsNotExist(err) {
		err = ioutil.WriteFile(fdFilename, []byte("{}"), 0777)
		if err != nil {
			return nil, err
		}
	}

	log.Println("start loading data from file...")
	err = s.tfidf.LoadFrom(pdFilename, fdFilename)
	return s, err
}

func (s *Server) Save(pdFilename, fdFilename string) error {
	return s.tfidf.Save(pdFilename, fdFilename)
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

func (s *Server) GetStatistics(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, struct {
		DocCount  int `json:"doc_count"`
		WordCount int `json:"word_count"`
	}{
		s.tfidf.DocCount(),
		s.tfidf.WordCount(),
	})
}
