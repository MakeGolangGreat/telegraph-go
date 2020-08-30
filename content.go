package telegraph

// Copy from https://github.com/toby3d/telegraph/blob/master/content.go
// and Modifiy something.

import (
	"bytes"
	"errors"
	"io"
	"strings"

	"golang.org/x/net/html"
)

// 将字符串HTML转换为HTML Node数据结构
func contentFormat(data interface{}) (n []Node, err error) {
	var dst *html.Node

	switch src := data.(type) {
	case string:
		dst, err = html.Parse(strings.NewReader(src))
	case []byte:
		dst, err = html.Parse(bytes.NewReader(src))
	case io.Reader:
		dst, err = html.Parse(src)
	default:
		return nil, errors.New("invalid data type")
	}

	if err != nil {
		return nil, err
	}

	switch node := domToNode(dst.FirstChild).(type) {
	case *NodeElement:
		// 在返回的时候，因为多了几层无效的结构（对应html-head/body），所以这里直接读取body下的children，然后返回。
		switch bodyChild := node.Children[1].(type) {
		case *NodeElement:
			n = bodyChild.Children
		}
	case nil:
		n = append(n, &NodeElement{
			Tag:      "p",
			Children: []Node{"没有内容"},
		})
	case string:
		n = append(n, &NodeElement{
			Tag:      "p",
			Children: []Node{node},
		})
	}

	return n, nil
}

// 递归解析DOM，返回Node数据
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
