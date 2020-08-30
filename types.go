package telegraph

import "fmt"

func init() {
	fmt.Println("types.go init...")
}

type (
	// Page 是创建文章时的请求类型，内含创建文章需要的请求数据
	Page struct {
		AccessToken   string      `json:"access_token" validate:"required"`
		Title         string      `json:"title" validate:"max=10,min=1,required"`
		Content       interface{} `json:"content" validate:"required"`
		ReturnContent string      `json:"return_content,omitempty"`
		Data          string      `json:"-"`
		Debug         bool        `json:"-"`
		AttachInfo    *NodeElement `json:"-"`
		AuthorName    string      `json:"author_name,omitempty"`
		AuthorURL     string      `json:"author_url,omitempty"`
	}

	// PageResponse 创建Page接口返回的数据
	PageResponse struct {
		OK     bool
		Result struct {
			Path string
			URL  string
		}
		Error string
	}

	// Node 对应DOM中的一个Node节点，可以是字符串或者NodeElement
	Node interface{}
	// NodeElement 表示一个元素，因为Telegraph的限制，只允许这三个属性
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
