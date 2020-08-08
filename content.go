// Copy from https://github.com/toby3d/telegraph/blob/master/content.go
package telegraph

import (
	"bytes"
	"errors"
	"io"
	"strings"

	"golang.org/x/net/html"
)

type (
	// Node 代表着html数据的Node数据
	Node interface{}
	// NodeElement 是节点元素
	NodeElement struct {
		// Name of the DOM element. Available tags: a, aside, b, blockquote, br, code, em, figcaption, figure,
		// h3, h4, hr, i, iframe, img, li, ol, p, pre, s, strong, u, ul, video.
		Tag string `json:"tag"`

		// Attributes of the DOM element. Key of object represents name of attribute, value represents value
		// of attribute. Available attributes: href, src.
		Attrs map[string]string `json:"attrs,omitempty"`

		// List of child nodes for the DOM element.
		Children []Node `json:"children,omitempty"`
	}
)

var ErrInvalidDataType = errors.New("invalid data type")

// ContentFormat transforms data to a DOM-based format to represent the content of the page.
func ContentFormat(data interface{}) (n []Node, err error) {
	var dst *html.Node

	switch src := data.(type) {
	case string:
		dst, err = html.Parse(strings.NewReader(src))
	case []byte:
		dst, err = html.Parse(bytes.NewReader(src))
	case io.Reader:
		dst, err = html.Parse(src)
	default:
		return nil, ErrInvalidDataType
	}

	if err != nil {
		return nil, err
	}

	n = append(n, domToNode(dst.FirstChild))

	return n, nil
}

func domToNode(domNode *html.Node) interface{} {
	if domNode.Type == html.TextNode {
		return domNode.Data
	}

	if domNode.Type != html.ElementNode {
		return nil
	}

	nodeElement := new(NodeElement)

	switch strings.ToLower(domNode.Data) {
	case "a", "aside", "b", "blockquote", "br", "code", "em", "figcaption", "figure", "h3", "h4", "hr", "i",
		"iframe", "img", "li", "ol", "p", "pre", "s", "strong", "u", "ul", "video":
		nodeElement.Tag = domNode.Data

		for i := range domNode.Attr {
			switch strings.ToLower(domNode.Attr[i].Key) {
			case "href", "src":
				nodeElement.Attrs = map[string]string{domNode.Attr[i].Key: domNode.Attr[i].Val}
			default:
				continue
			}
		}
	}

	for child := domNode.FirstChild; child != nil; child = child.NextSibling {
		nodeElement.Children = append(nodeElement.Children, domToNode(child))
	}

	return nodeElement
}
