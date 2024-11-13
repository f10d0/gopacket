# TochusC/GoPacket
该库为`google/gopacket`的分支版本，除将部分文档进行中文化以外，特别地对layers/dns.go文件进行了修改，以进行DNS相关实验。

## 引入的新特性🌟
- 中文化的文档介绍！
- 支持构建并发送未知类型的Resource Recourd（RR），
- 添加对DNSSEC相关RR的基本支持

## 未来工作🛠️
- 继续添加更多中文文档
- 进一步扩展对DNSSEC的支持
- 添加更多RR类型

有关更多详细信息，请参阅 [godoc](https://godoc.org/github.com/google/gopacket)。

[![GoDoc](https://godoc.org/github.com/google/gopacket?status.svg)](https://godoc.org/github.com/google/gopacket)

最低 Go 版本要求是 1.9

最初从 Andreas Krennmair <ak@synflood.at> (http://github.com/akrennmair/gopcap) 编写的 gopcap 项目中分叉出来。

`layers\dns.go`中的相关代码参考了 `gopacket\gopacket`(http://github.com/gopacket/gopacket) 项目，感谢社区贡献！
