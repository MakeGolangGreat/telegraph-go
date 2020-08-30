package telegraph

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/fatih/color"
)

func init() {
	// log.SetFlags(log.LstdFlags | log.Lshortfile)
	fmt.Println("telegraph.go init...")
}

// 启动Debug模式，才会打印错误
func (page *Page) logError(msg string, err error) {
	if page.Debug {
		color.Red("$s: $s", msg, err.Error())
	}
}

// SendPage 发送文章请求
func (page *Page) SendPage(url string, client *http.Client) (link string, err error) {
	payload, err := json.Marshal(page)
	if err != nil {
		page.logError("文章数据字符化失败", err)
		return "", err
	}

	request, newRequestErr := http.NewRequest(http.MethodPost, url, strings.NewReader(string(payload)))
	if newRequestErr != nil {
		page.logError("NewRequest 失败", newRequestErr)
		return "", newRequestErr
	}
	request.Header.Set("Content-Type", "application/json")

	resp, doError := client.Do(request)
	if doError != nil {
		page.logError("发起请求失败", doError)
		return "", doError
	}
	defer resp.Body.Close()

	body, readError := ioutil.ReadAll(resp.Body)
	if readError != nil {
		page.logError("读取resp.body失败", readError)
		return "", readError
	}

	var pageResponse PageResponse
	unmarshalError := json.Unmarshal(body, &pageResponse)
	if unmarshalError != nil {
		page.logError("反字符化失败", unmarshalError)
		return "", unmarshalError
	}

	if !pageResponse.OK {
		// TODO 自定义Error没实现Error()方法，导致抛出这个错误时看不到错误字符串。
		return "", errors.New(pageResponse.Error)
	}

	return pageResponse.Result.URL, nil
}

// CreatePage 将文本数据保存到Telegrah，并返回一个Telegraph链接。
// 具体实现和说明请看：CreatePageWithClient
// 这里只是一层简单的封装，传入一个默认的Client
func (page *Page) CreatePage() (string, error) {
	return page.CreatePageWithClient(&http.Client{})
}

// CreatePageWithClient 将文本数据保存到Telegrah，并返回一个Telegraph链接。
// 由于Telegraph存在内容大小限制（64kb,utf-8），CreatePageWithClient将会把文本数据分割成多篇文章发表。
// 但放心，第一篇文章尾部将会存在指向第二篇文章的链接（「下一页」）
// 提供一个Client参数，方便调用者传入一个自定义的Client（其实就是代理，让程序可以翻墙）
func (page *Page) CreatePageWithClient(client *http.Client) (string, error) {
	pageNode, err := contentFormat(page.Data)
	if err != nil {
		page.logError("DOM -> Node失败", err)
		return "", err
	}

	pageStr, marshalError := json.Marshal(pageNode)
	if marshalError != nil {
		page.logError("page字符化失败", marshalError)
		return "", marshalError
	}

	// 全部文章数据
	// 这里还没有根据Telegraph能发的最大数据限制来分割数据。
	var totalArticleNodeArray []Node
	// 如果数据字节够一篇文章
	if int32(len(pageStr)) < maxContentLimit {
		totalArticleNodeArray = append(totalArticleNodeArray, pageNode)
	} else {
		// 如果数据字节超过一篇文章，那么需要遍历将文章数据分割成多篇文章。
		var byteSizeCount int32 // 字节计数
		var index int           // []Node数据的下标索引

		for i, v := range pageNode {
			// 判断字节时，必须将[]Node数据转换为字符串。
			payload, payloadError := json.Marshal(v)
			if payloadError != nil {
				page.logError("字符化失败", payloadError)
				return "", payloadError
			}

			// 每次遍历，累加遍历过的元素的字节数。
			byteSizeCount += int32(len(payload))

			// 如果字节大小超过限制，说明算上当前下标对应的数据，已经超过一篇文章的字节大小。
			// 那么到前一个下标为止的数据量够一篇文章了。
			if byteSizeCount > maxContentLimit {
				totalArticleNodeArray = append(totalArticleNodeArray, pageNode[index:i])
				// 重新设置起点
				index = i
				// 重置字节计数器
				byteSizeCount = 0
			}
		}

		// 最后把上个循环最后没有合并进来的尾部元素单独合并进来。
		// 一定有最后一部分内容等待手动合并。
		totalArticleNodeArray = append(totalArticleNodeArray, pageNode[index:])
	}

	// 记录上一个请求返回的链接，也就是「下一页」的链接。
	var previousArticleLink string
	// 文章数据。分割后的，一篇Telegraph文章的数据。
	var articleNodeArray []Node
	// 倒序发出请求，为了在当前文章中添加「下一页」
	for i := len(totalArticleNodeArray) - 1; i >= 0; i-- {
		switch nodeArr := totalArticleNodeArray[i].(type) {
		case []Node:
			articleNodeArray = nodeArr

			// 如果存在链接，说明一共不只有一篇文章，且这不是倒序发出的第一篇，需要手动添加上「下一页」
			if previousArticleLink != "" {
				articleNodeArray = append(articleNodeArray, NodeElement{
					Tag: "p",
					Children: []Node{
						NodeElement{
							Tag: "br",
						},
						NodeElement{
							Tag: "a",
							Attrs: map[string]string{
								"href": previousArticleLink,
							},
							Children: []Node{"下一页"},
						},
					},
				})
			}

			// 给每篇Telegraph文章的下方都添加上项目的链接。
			articleNodeArray = append(articleNodeArray, page.AttachInfo)
		}

		content, contentError := json.Marshal(articleNodeArray)
		if contentError != nil {
			page.logError("文章数据字符化失败", contentError)
			return "", contentError
		}

		contentStr := string(content)
		if page.Debug {
			color.Red("当前文章序号：%d\n，文章字节数：%d", i+1, len(contentStr))
		}

		page.Content = contentStr

		link, SendPageError := page.SendPage(createPageURL, client)
		if SendPageError != nil {
			page.logError("createPage 请求发送失败", SendPageError)
			return "", SendPageError
		}

		previousArticleLink = link
	}

	return previousArticleLink, nil
}
