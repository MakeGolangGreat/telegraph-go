package telegraph

const (
	maxContentLimit int32    = 60000 // 接口限制字节长度为64KB，utf-8编码，所以这里用60000。毕竟还得考虑数据中最后还添加了下一页/项目介绍这样的数据
	createPageURL   string = "https://api.telegra.ph/createPage"
)
