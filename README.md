# telegraph-go

> 是一个封装了 [Telegraph API](https://telegra.ph/api) 的`Golang`库。
> 
> 传入`HTML字符串`和`标题`数据（当然`telegraph-token`是必须的），telegraph-go会将它存储到Telegraph上，并返回一个可访问的链接给你。
> 
> 如何获取`telegraph-token`？请看上面 [Telegraph API](https://telegra.ph/api) 文档

### APIs has achieved

---

- [x] `CreatePage`

### Getting Started

---

1. Download

```go
go get -u github.com/MakeGolangGreat/telegraph-go
```

2. `test.go`

```go
package main

imoprt "github.com/MakeGolangGreat/telegraph-go"

func main(){
  page := &telegraph.Page{
    AccessToken: "......<telegraph-token>......",
		AuthorURL:   "https://github.com/MakeGolangGreat/telegraph-go",
		AuthorName:  "telegraph-go",
    Title: 			 "Title here",
    Data:				 "<h1>Put html strings here.</h1>",
	}
  
  link, err := page.CreatePage()
	if err != nil {
    fmt.Println("Create Page Failed: ", err)
  }else{
    fmt.Println(link) 
  }
}
```

It's a very simple sample above. You maybe want to look [this archive project](https://github.com/MakeGolangGreat/archive-go) that using `telegraph-go` now.

