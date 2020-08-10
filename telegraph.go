package telegraph

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

const (
	createPageURL string = "https://api.telegra.ph/createPage"
)

var validate *validator.Validate

type CreatePageRequest struct {
	AccessToken   string `json:"access_token"`
	Title         string `json:"title" validate:"max=10,min=1"`
	Content       []Node `json:"content"`
	ReturnContent string `json:"return_content"`
	Data          string `json:-`
	AuthorName    string `json:"author_name"`
	AuthorURL     string `json:"author_url"`
}

// 暂时只有自己用得到的属性
type Page struct {
	Path string
	URL  string
}

// 创建Page接口返回的数据
type ResData struct {
	OK     bool
	Result Page
	Error  string
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func errorHandler(msg string, err error) {
	if err != nil {
		log.Fatal("%s - %s", msg, err.Error)
	}
}

func sendPost(url string, data *CreatePageRequest, client *http.Client) (link string, err error) {
	payload, err := json.Marshal(data)
	errorHandler("JSON化失败", err)

	request, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(payload)))
	request.Header.Set("Content-Type", "application/json")
	errorHandler("初始化请求失败", err)

	resp, err2 := client.Do(request) // resp 可能为 nil，不能读取 Body。所以不能先defer
	errorHandler("执行请求失败", err2)
	defer resp.Body.Close()

	body, err3 := ioutil.ReadAll(resp.Body)
	errorHandler("ioutil.ReadAll 失败", err3)

	var dataStruct ResData
	err4 := json.Unmarshal(body, &dataStruct)
	errorHandler("解析Response失败", err4)

	// 上面内容长度没有把握好造成的
	// 如果为false，十有八九是因为内容过长：CONTENT_TOO_BIG
	if !dataStruct.OK {
		// TODO 自定义Error没实现Error()方法，导致抛出这个错误时看不到错误字符串。
		return "", errors.New(dataStruct.Error)
	}

	return dataStruct.Result.URL, nil
}

// CreatePage 保存文章，支持判断文章字符是否超出 64 * 1024 byte/字节，超出的话自动截取并且生成多篇文章，且自动将第二篇文章的链接添加到第一篇文章里。
func CreatePage(data *CreatePageRequest) (string, error) {
	// 接口限制字节长度为64KB，utf-8编码，意味着全是中文字符的话，差不多只能有 21000 个字。
	// 很奇怪：64 * 1024 / 3 = 21845，但我看Request Content-Length是 66318。多出来的，2318kb是什么数据。猜测是请求的其他部分的内容。懒得查了，大概以21000个字符为上限好了。

	// 但为了保留可能存在的后面内容的telegra.ph链接，预留出部分字节长度。
	const limit int32 = 21000 //21000
	// [0 21000 25000]
	var pageSlice = []int32{0}
	// 转换位数组，方便获取子字符串
	contentRune := []rune(data.Data)

	// 根据limit将内容分割成下标的数组
	fmt.Println("文章的长度是: ", int32(len(contentRune)))

	for i := int32(0); i < int32(len(contentRune)); {
		i += limit
		pageSlice = append(pageSlice, i)
	}

	fmt.Println(pageSlice)

	// 然后倒着保存文章，这是为了先获取到后面文章的链接。
	// 从切片倒数第二个开始遍历

	// 存储下一篇文章的链接。
	var nextPageLink string

	for i := len(pageSlice) - 2; i >= 0; i-- {
		// 截取的文章字符串
		var page string
		// 获取到对应下标范围内的文章字符串。
		if i == len(pageSlice)-2 {
			// 获取到对应下标范围内的文章字符串。
			page = string(contentRune[pageSlice[i]:])
		} else {
			page = string(contentRune[pageSlice[i]:pageSlice[i+1]])
		}

		if nextPageLink != "" {
			page += fmt.Sprintf("\n\n<a href='%s'>下一页</a>", nextPageLink)
		}

		content, err := ContentFormat(page)
		errorHandler("转换文章为[]Node失败", err)

		data.Content = content

		link, err2 := sendPost(createPageURL, data, &http.Client{})
		errorHandler("保存文章失败", err2)
		nextPageLink = link
	}

	return nextPageLink, nil
}

func CreatePageWithClient(data CreatePageRequest, client *http.Client) (string, error) {
	content, err := ContentFormat(data.Data)
	errorHandler("转换文章数据失败", err)
	data.Content = content

	return sendPost(createPageURL, &data, client)
}
