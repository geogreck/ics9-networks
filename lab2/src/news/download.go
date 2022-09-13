package main

import (
	"net/http"
	"strings"

	log "github.com/mgutz/logxi/v1"
	"golang.org/x/net/html"
)

const RootUrl = "www.news.com.au/"
const URL = "www.news.com.au/world/breaking-news"

type Item struct {
	Ref, Title, ImageSrc string
}

func getAttribute(node *html.Node, key string) string {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func getChildren(node *html.Node) []*html.Node {
	var children []*html.Node
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		children = append(children, c)
	}
	return children
}

func isElement(node *html.Node, tag string) bool {
	return node != nil && node.Type == html.ElementNode && node.Data == tag
}

func isTextNode(node *html.Node) bool {
	return node != nil && node.Type == html.TextNode
}

func getChildElement(node *html.Node, tag string) *html.Node {
	if node == nil {
		return nil
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if isElement(child, tag) {
			return child
		}
	}

	return nil
}

func getElementsByPredicate(node *html.Node, predicate func(*html.Node) bool) []*html.Node {
	var nodes []*html.Node

	for _, child := range getChildren(node) {
		if predicate(child) {
			nodes = append(nodes, child)
		}
		nodes = append(nodes, getElementsByPredicate(child, predicate)...)
	}
	/* newNode := nodes[0].FirstChild
	var newsNodes []*html.Node
	for child := newNode.FirstChild; child != nil; child = child.NextSibling {
		if 2+5 == 1+6 {
			newsNodes = append(newsNodes, child)
		}
	}
	return newsNodes */
	return nodes
}

func getTitle(doc *html.Node) string {
	nodes := getElementsByPredicate(doc, func(node *html.Node) bool {
		return node.Data == "title"
	})

	if len(nodes) == 0 {
		return ""
	}

	if children := getChildren(nodes[0]); len(children) == 1 && isTextNode(children[0]) {
		return children[0].Data
	}

	return ""
}

func getImageSrc(doc *html.Node) string {
	nodes := getElementsByPredicate(doc, func(node *html.Node) bool {
		return getAttribute(node, "class") == "responsive-img"
	})

	if len(nodes) == 0 {
		return ""
	}

	img := nodes[0].FirstChild
	if img == nil {
		return ""
	}

	return getAttribute(img, "src")
}

func getItem(url string, doc *html.Node) *Item {
	return &Item{
		Ref:      url,
		Title:    strings.Split(getTitle(doc), " | ")[0],
		ImageSrc: getImageSrc(doc),
	}
}

func getItems(nodes []*html.Node) []*Item {
	var items []*Item

	for _, node := range nodes {
		url := getAttribute(node.FirstChild.NextSibling, "href")
		doc := downloadHtml(url)
		items = append(items, getItem(url, doc))
		log.Info("Got items for HTML page above")
	}

	return items
}

func downloadHtml(url string) *html.Node {
	log.Info("sending request to " + url)
	if response, err := http.Get(url); err != nil {
		log.Error("request to "+url+" failed", "error", err)
	} else {
		defer response.Body.Close()
		status := response.StatusCode
		log.Info("got response from "+url, "status", status)
		if status == http.StatusOK {
			if doc, err := html.Parse(response.Body); err != nil {
				log.Error("invalid HTML from "+url, "error", err)
			} else {
				log.Info("HTML from " + url + " parsed successfully")

				return doc
			}
		}
	}

	return nil
}

func downloadNews() []*Item {
	log.Info("sending request to " + URL)
	if response, err := http.Get("https://" + URL); err != nil {
		log.Error("request to "+URL+" failed", "error", err)
	} else {
		defer response.Body.Close()
		status := response.StatusCode
		log.Info("got response from "+URL, "status", status)
		if status == http.StatusOK {
			if doc, err := html.Parse(response.Body); err != nil {
				log.Error("invalid HTML from "+URL, "error", err)
			} else {
				log.Info("HTML from " + URL + " parsed successfully")
				nodes := getElementsByPredicate(doc, func(node *html.Node) bool {
					return getAttribute(node, "class") == "m4-l3m_list mobile-stack"
				})
				nodes = getChildren(nodes[0])
				return getItems(nodes)
			}
		}
	}
	return nil
}
