package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/Sudalight/tools/pkg/convert"
)

var (
	cookie    = flag.String("c", "", "cookie for request in gitlab")
	groupList = flag.String("groups", "", "groups to parse, separate by ,")
	host      = flag.String("host", "", "host for gitlab")
	cli       = http.Client{}
	groups    = []string{}
)

const (
	schema = "https"
)

type GoModHTML struct {
	HTML string `json:"html"`
}

type GroupChild struct {
	Type         string `json:"type"`
	ProjectCount int    `json:"project_count"`
	RelativePath string `json:"relative_path"`
}

type Project struct {
	Path            string
	DependencyPaths []string
}

type ProjectNode struct {
	Module       string
	Dependencies []*ProjectNode
}

type Tree struct {
	Nodes      []Node     `json:"nodes"`
	Links      []Link     `json:"links"`
	Categories []Category `json:"categories"`
}

type Node struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	SymbolSize float64 `json:"symbolSize"`
	X          float64 `json:"x"`
	Y          float64 `json:"y"`
	Value      float64 `json:"value"`
	Category   int     `json:"category"`
}

type Link struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type Category struct {
	Name string `json:"name"`
}

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	flag.Parse()
	if *host == "" {
		panic("host must be provided")
	}
	if *cookie == "" {
		panic("cookie must be provided")
	}
	if *groupList == "" {
		panic("groupList must be provided")
	}
	groups = strings.Split(*groupList, ",")

	var projectPaths []string
	for i := range groups {
		paths, err := getProjectPaths(groups[i], 1)
		if err != nil {
			log.Println(groups[i], err)
		}
		projectPaths = append(projectPaths, paths...)
	}

	var projects []Project
	for i := range projectPaths {
		dependencyPaths, err := getDependency(projectPaths[i])
		if err != nil {
			log.Println(projectPaths[i], err)
			continue
		}
		projects = append(projects, Project{
			Path:            projectPaths[i],
			DependencyPaths: dependencyPaths,
		})
	}

	projectMap := make(map[string]*ProjectNode)
	for i := range projects {
		node := &ProjectNode{
			Module: *host + projects[i].Path,
		}
		projectMap[node.Module] = node
	}

	nodes := make([]Node, 0)
	links := make([]Link, 0)
	categories := make([]Category, 0)
	categorySet := make(map[string]int)
	for i := range projects {
		index, ok := categorySet[strings.TrimPrefix(projects[i].Path, *host)[:strings.LastIndex(projects[i].Path, "/")]]
		if !ok {
			categories = append(categories, Category{
				Name: strings.TrimPrefix(projects[i].Path, *host)[:strings.LastIndex(projects[i].Path, "/")],
			})
			categorySet[strings.TrimPrefix(projects[i].Path, *host)[:strings.LastIndex(projects[i].Path, "/")]] = len(categories) - 1
			index = len(categories) - 1
		}
		nodes = append(nodes, Node{
			ID:         *host + projects[i].Path,
			Name:       *host + projects[i].Path,
			SymbolSize: 20,
			Value:      10,
			Category:   index,
			X:          float64(rand.Intn(40)),
			Y:          float64(rand.Intn(40)),
		})
		for j := range projects[i].DependencyPaths {
			if node, ok := projectMap[projects[i].DependencyPaths[j]]; !ok {
				log.Printf("%s requires %s, not found", projects[i].Path, projects[i].DependencyPaths[j])
			} else {
				links = append(links, Link{
					Source: *host + projects[i].Path,
					Target: node.Module,
				})
			}
		}
	}
	data, err := json.Marshal(Tree{
		Nodes:      nodes,
		Links:      links,
		Categories: categories,
	})
	if err != nil {
		panic(err)
	}
	ioutil.WriteFile("data.json", data, 0644)
}

func getProjectPaths(group string, page int) ([]string, error) {
	req := &http.Request{
		Method: http.MethodGet,
		URL:    newProjectsURL(group, page),
		Header: http.Header{
			"Cookie": []string{*cookie},
		},
	}
	res, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, err
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	children := make([]GroupChild, 0)
	err = json.Unmarshal(body, &children)
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0)
	for i := range children {
		if children[i].Type == "project" {
			paths = append(paths, children[i].RelativePath)
		} else if children[i].ProjectCount > 0 {
			subPaths, err := getProjectPaths(children[i].RelativePath, 1)
			if err != nil {
				return nil, err
			}
			paths = append(paths, subPaths...)
		}
	}

	totalPages, err := strconv.Atoi(res.Header.Get("X-Total-Pages"))
	if err != nil {
		return nil, err
	}
	if page < totalPages {
		nextPagePaths, err := getProjectPaths(group, page+1)
		if err != nil {
			return nil, err
		}
		paths = append(paths, nextPagePaths...)
	}
	return paths, nil
}

func getDependency(path string) ([]string, error) {
	req := &http.Request{
		Method: http.MethodGet,
		URL:    newGoModURL(path, "dev"),
		Header: http.Header{
			"Cookie": []string{*cookie},
		},
	}
	res, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNotFound {
		return nil, err
	} else if res.StatusCode == http.StatusNotFound {
		req.URL = newGoModURL(path, "master")
		res, err = cli.Do(req)
		if err != nil {
			return nil, err
		}
		if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNotFound {
			return nil, errors.New(res.Status)
		} else if res.StatusCode == http.StatusNotFound {
			return nil, errors.New("not go project")
		}
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	a := GoModHTML{}
	err = json.Unmarshal(body, &a)
	if err != nil {
		log.Println(res.StatusCode, path)
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewBuffer(convert.StringToBytes(a.HTML)))
	if err != nil {
		return nil, err
	}
	aDoc := doc.Find("span")
	dependencies := make([]string, 0)
	aDoc.Each(func(i int, s *goquery.Selection) {
		// For each item found, get the dependency
		dependency := s.Find("a").Text()
		if dependency == "" || !strings.HasPrefix(dependency, *host) ||
			dependency == *host+path {
			return
		}
		dependencies = append(dependencies, dependency)
	})
	return dependencies, nil
}

func newProjectsURL(group string, page int) *url.URL {
	params := url.Values{}
	if page > 1 {
		params.Add("page", strconv.Itoa(page))
	}
	return &url.URL{
		Scheme:   schema,
		Host:     *host,
		Path:     fmt.Sprintf("/groups%s/-/children.json", group),
		RawQuery: params.Encode(),
	}
}

func newGoModURL(path, branch string) *url.URL {
	params := url.Values{}
	params.Add("format", "json")
	params.Add("viewer", "simple")
	return &url.URL{
		Scheme:   schema,
		Host:     *host,
		Path:     fmt.Sprintf("%s/-/blob/%s/go.mod", path, branch),
		RawQuery: params.Encode(),
	}
}
