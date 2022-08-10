/*
Returns: 0 on success, 1 on failure (no results)
Outputs: if found, output location and info



*/

package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

func getHtml(url string, client *http.Client) (raw string, err error) {
	var resp *http.Response
	if resp, err = client.Get(url); err != nil {
		fmt.Printf("Error %s", err)
		return
	}
	defer resp.Body.Close()

	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return
	}

	raw = string(body)
	return
}

func parsePage(page string) (doc *html.Node, err error) {
	doc, err = html.Parse(strings.NewReader(page))
	if err != nil {
		log.Fatal(err)
	}
	return
}

func renderNode(n *html.Node) string {
	var buf bytes.Buffer
	w := io.Writer(&buf)

	err := html.Render(w, n)
	if err != nil {
		return ""
	}

	return buf.String()
}

func parsePullAndSave(page string) (r *result) {
	var doc *html.Node
	doc, _ = parsePage(page)

	foundList := ""
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "span" && n.FirstChild != nil {
			for _, attr := range n.Attr {
				if attr.Key == "id" && strings.Contains(attr.Val, "yard_locations_Year") {
					s := strings.Split(renderNode(n), "<span>")[1]
					s = strings.Split(s, "</span>")[0]
					if year, err := strconv.Atoi(s); err == nil {
						//fmt.Println("HIT", year)
						if year >= 1984 && year <= 1988 {
							foundList = fmt.Sprintf("%s %d", foundList, year)
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	if foundList != "" {
		r = &result{
			description: foundList,
		}
	}

	return
}

type result struct {
	description string
	location    string
}

type source struct {
	url     string
	parseFn func(string) *result
}

var (
	sources = []source{
		{
			url:     "https://newautopart.net/includes/pullandsave/spokane/yard_locationslist.php?cmd=search&t=yard_locations&psearch=maxima&psearchtype=",
			parseFn: parsePullAndSave,
		},
	}
)

func main() {
	cli := &http.Client{Timeout: time.Duration(5) * time.Second}

	var err error
	var page string
	if page, err = getHtml(sources[0].url, cli); err != nil {
		fmt.Println(err)
		return
	}

	if result := parsePullAndSave(page); result != nil {
		fmt.Println("Found: ")
		fmt.Println(result.description)
		return
	}

	fmt.Println("nothing found")
	return
}
