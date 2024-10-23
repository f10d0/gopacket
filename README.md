# TochusC/GoPacket
该库为`google/gopacket`的分支版本，除为 Go 提供了数据包解码功能外，特别地对layers/dns.go文件进行了修改，以进行DNS相关实验。

## 引入的新特性🌟
- 支持构建并发送未知类型的Resource Recourd（RR），
- 添加对DNSSEC相关RR的基本支持

## 未来工作🛠️
- 进一步扩展对DNSSEC的支持
- 添加更多RR类型

有关更多详细信息，请参阅 [godoc](https://godoc.org/github.com/tochusc/gopacket)。

[![Build Status](https://travis-ci.org/google/gopacket.svg?branch=master)](https://travis-ci.org/google/gopacket)
[![GoDoc](https://godoc.org/github.com/tochusc/gopacket?status.svg)](https://godoc.org/github.com/tochusc/gopacket)

最低 Go 版本要求是 1.9

最初从 Andreas Krennmair <ak@synflood.at> (http://github.com/akrennmair/gopcap) 编写的 gopcap 项目中分叉出来。

`layers\dns.go`中的相关代码参考了 `gopacket\gopacket`(http://github.com/gopacket/gopacket) 项目，感谢他们的贡献！
