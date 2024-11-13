// Copyright 2024 TochusC AOSP Lab. All rights reserved.
// Copyright 2012 Google, Inc. All rights reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file in the root of the source
// tree.

/*
layers 包提供了许多常见协议的 待解码层（Decoder Layers） 实现。

layers 包实现了许多常见协议层的解码功能，
使用 gopacket 的人几乎总是需要用到 layers 来将数据包解码为有用的部分。
要了解 gopacket/layers 目前能够解码的协议集合，请查看 Variables 部分定义的 LayerTypes 集合。
layers 包还为常见的，拥有源地址、目的地址的数据包协议层定义了 端点（Endpoints），
如 IPv4/6（IPs）和 TCP/UDP（端口）。
最后，layers 包含了许多有用的枚举类型（IPProtocol、EthernetType、LinkType、PPPType 等）。
其中许多都实现了 gopacket.Decoder 接口，因此可以将它们作为 待解码层 传递给 gopacket。

大多数常见的协议层使用首字母缩写或其他行业常见名称（IPv4、TCP、PPP）命名。
一些不太常见的协议层则将其名称展开（CiscoDiscoveryProtocol）。
对于某些协议，协议的子部分被拆分到自己的层中（例如 SCTP）。
这主要是在协议的某些部分可能包含了 内容层（Interesting Layers），
（SCTPData 实现 ApplicationLayer，而基本 SCTP 实现 TransportLayer），
或者可能是因为将协议拆分为几个层将使解码更容易等。

这个包是为了与其父包 http://github.com/google/gopacket 一起使用而设计的。

# 端口类型

gopacket 使用不同的端口类型来代替原始的 uint16 或 uint8 值，例如 TCPPort 和 UDPPort。
而不是使用原始的 uint16 或 uint8 值来表示端口，
这使得 gopacket 可以为每个端口重写字符串行为，
gopacket 通过设置端口名称映射（TCPPortNames、UDPPortNames 等）来实现这一点。
知名端口使用其协议名称进行注释，并且其 String 函数显示这些名称：

	p := TCPPort(80)
	fmt.Printf("Number: %d  String: %v", p, p)
	// Prints: "Number: 80  String: 80(http)"

# 修改解码行为

layers 包通过枚举将解码链接在一起。例如，在解码以太网层类型之后，
它使用 Ethernet.EthernetType 作为下一个解码器。
所有充当解码器的枚举类型，如 EthernetType，都可以根据用户的喜好进行修改。
例如，如果您有一个比 layers 包内置的 IPv4 解码器更好得多的新 IPv4 解码器，
您可以这样做：

	var mySpiffyIPv4Decoder gopacket.Decoder = ...
	layers.EthernetTypeMetadata[EthernetTypeIPv4].DecodeWith = mySpiffyIPv4Decoder

这将使所有未来的以太网数据包都使用您的新解码器来解码 IPv4 数据包，
而不是使用 gopacket 的内置解码器。
*/
package layers
