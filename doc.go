// Copyright 2024 TochusC, Inc. All rights reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file in the root of the source
// tree.

/*
gopacket 提供了用于 Go 语言的数据包解码功能。

gopacket 包含许多子包，这些子包提供了额外的功能，包括：

  - layers：您可能每次都将会使用layers。它包含了 gopacket 内置的用于解码数据包协议的逻辑。请注意，以下的所有示例代码都将假设您已经导入了 gopacket 和 gopacket/layers。
  - pcap：C 绑定，用于使用 libpcap 从网络中读取数据包。
  - pfring：C 绑定，用于使用 PF_RING 从网络中读取数据包。
  - afpacket：C 绑定，用于使用 Linux 的 AF_PACKET 从网络中读取数据包。
  - tcpassembly：TCP 流重组

此外，如果您想直接进入代码，请参阅 examples 子目录，其中包含了许多使用 gopacket 库构建的简单二进制文件。

最低 Go 版本要求是 1.5，但 pcapgo/EthernetHandle、afpacket 和 bsdbpf 由于依赖 x/sys/unix ，所以至少需要 1.7 及以上的版本。

基本用法

gopacket 接收一个 []byte 类型的数据包数据，并将其解码为包含一个或多个“层”的数据包。
每一层对应于字节中的一个协议。一旦数据包被解码，数据包的层就可以直接从数据包中请求。

	// 解码一个数据包
	packet := gopacket.NewPacket(myPacketData, layers.LayerTypeEthernet, gopacket.Default)
	// 从这个数据包中获取 TCP 层
	if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
	 fmt.Println("这是一个 TCP 数据包！")
	 // 从该层获取实际的 TCP 数据
	 tcp, _ := tcpLayer.(*layers.TCP)
	 fmt.Printf("从源端口 %d 到目标端口 %d\n", tcp.SrcPort, tcp.DstPort)
	}
	// 遍历所有层，打印出每一层的类型
	for _, layer := range packet.Layers() {
	 fmt.Println("数据包层：", layer.LayerType())
	}

数据包可以从多个起点解码。我们的许多基本类型实现了 Decoder 接口，
这使我们能够解码我们没有完整数据的数据包。

	// 解码一个以太网数据包
	ethP := gopacket.NewPacket(p1, layers.LayerTypeEthernet, gopacket.Default)
	// 解码一个 IPv6 头及其包含的所有内容
	ipP := gopacket.NewPacket(p2, layers.LayerTypeIPv6, gopacket.Default)
	// 解码一个 TCP 头及其负载
	tcpP := gopacket.NewPacket(p3, layers.LayerTypeTCP, gopacket.Default)

从源读取数据包

大多数时候，您不会仅仅只拥有一个 []byte 类型的数据包数据。
相反，您可能会希望从某个地方（文件、接口等）读取数据包并处理它们。
为此，您需要构建一个 PacketSource 对象。

为此，您需要首先构造一个实现 PacketDataSource 接口的对象。
gopacket 在 gopacket/pcap 和 gopacket/pfring 子包中包含了这个接口的实现……
请参阅它们的文档以获取更多使用信息。
一旦您有了一个 PacketDataSource，您可以将其传递给 NewPacketSource，
并选择一个 Decoder 来创建一个 PacketSource。

一旦您有了一个 PacketSource 后，您可以通过多种方式从中读取数据包。
请参阅 PacketSource 的文档以获取更多详细信息。
最简单的方法是 Packets 函数，它返回一个通道，然后异步地将新数据包写入该通道，
如果 packetSource 遇到文件结束，则关闭该通道。

	packetSource := ...  // 使用 pcap 或 pfring 构造
	for packet := range packetSource.Packets() {
	  handlePacket(packet)  // 对每个数据包执行某些操作
	}

您可以通过设置 packetSource.DecodeOptions 中的字段来更改 packetSource 的解码选项……
请参阅以下部分以获取更多详细信息。


# 延迟解码

gopacket 可以选择性地延迟解码数据包数据，这代表gopacket
将只会在需要处理函数调用时才解码数据包层。

	// 创建一个数据包，但实际上还没有解码任何内容
	packet := gopacket.NewPacket(myPacketData, layers.LayerTypeEthernet, gopacket.Lazy)
	// 现在，解码数据包直到发现第一个 IPv4 层，但不再进一步解码。
	// 如果没有发现 IPv4 层，gopacket将解码整个数据包以寻找IPv4层。
	ip4 := packet.Layer(layers.LayerTypeIPv4)
	// 解码所有层并返回它们。在第一个 IPv4 层前已经被解码过的内容将不会被再次解码。
	layers := packet.Layers()

延迟解码的数据包不是并发安全的。
由于没有一次性解码所有层，所以每次调用 Layer() 或 Layers() 都因为要解码下一层而改变数据包。
如果一个数据包在多个 goroutine 中被并发使用，请不要使用 gopacket.Lazy 参数。
这样 gopacket 将完全解码数据包，使得后续函数调用不会改变数据包对象。


无拷贝解码

默认情况下，gopacket 将复制传递给 NewPacket 的切片参数，并将拷贝存储在生成的数据包中，
因此对原始切片的后续更改不会影响新生成的数据包及其所包含的层。
如果您可以保证底层切片不会被更改，您可以将 NoCopy 参数传递给 gopacket.NewPacket，
这将让 NewPacket 使用传入的切片本身，而不是使用拷贝来承载数据。

	// 这个通道返回新的字节切片，每个切片都指向一个新的内存位置，
 	// 该位置在数据包的持续时间内是不可变的。
	for data := range myByteSliceChannel {
	  p := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.NoCopy)
	  doSomethingWithPacket(p)
	}

最快的解码方法是同时使用 Lazy 和 NoCopy，但请注意上述的许多警告，对于某些实现，使用Lazy或NoCopy可能产生危险。


指向已知层的指针

在解码过程中，某些层被存储为数据包中的已知层类型。例如，IPv4 和 IPv6 都被视为网络层（NetworkLayer）层，
而 TCP 和 UDP 都被视为传输层（TransportLayer）层。gopacket采用 4 层划分，分别对应于 TCP/IP 分层方案的 4 层
（大致类似于 OSI 模型的第 2、3、4 和 7 层）。要访问这些层，可以使用 packet.LinkLayer、packet.NetworkLayer、
packet.TransportLayer 和 packet.ApplicationLayer 函数。每个函数返回一个相应的接口（gopacket.{Link,Network,Transport,Application}Layer）。
前三层（LinkLayer, NetworkLayer, TransportLayer）的接口提供了获取该特定层的源/目标地址的方法，而最后一层（ApplicationLayer）的接口则提供 Payload 函数以获取其有效负荷。
这可以帮助您忽视底层数据类型，从而直接获取所有数据包的有效负荷：

// 从某个源获取数据包

	for packet := range someSource {
	  if app := packet.ApplicationLayer(); app != nil {
	    if strings.Contains(string(app.Payload()), "magic string") {
	      fmt.Println("在数据包中找到了奇妙字符串（magic string）!")
	    }
	  }
	}

一个特别有用的层是 ErrorLayer，当数据包某部分出现解码错误时会设置该层。

	packet := gopacket.NewPacket(myPacketData, layers.LayerTypeEthernet, gopacket.Default)
	if err := packet.ErrorLayer(); err != nil {
	  fmt.Println("解码数据包某部分时发生错误:", err)
	}

请注意 gopacket 并没有从 NewPacket 中返回错误，
这是因为在遇到错误的层之前，gopacket 可能已经成功解码了许多层。
这使得即使数据包的 TCP 层可能格式不正确，但您仍然可以正确获取数据包的 以太网 和 IPv4 层数据。


流和端点

gopacket 有着两个重要的概念：流和结束点。
在忽视协议的情况下，信息传输实质是从A到B的数据包流。
常用的层接口 LinkLayer、NetworkLayer 和 TransportLayer
都提供了能够忽视底层，而直接提取该层数据流信息的方法。


一个 流（Flow） 是一个由两个端点（Endpoint）组成的简单对象，一个源端点和一个目标端点。
它详细描述了数据包在该层的发送者和接收者。


一个端点（Endpoint）是源或目标的可哈希表示。
例如，对于 LayerTypeIPv4，一个端点包含了 v4 IP 数据包的 IP 地址字节。
一个流可以被分解为端点，并且端点可以被组合成流。

	packet := gopacket.NewPacket(myPacketData, layers.LayerTypeEthernet, gopacket.Lazy)
	netFlow := packet.NetworkLayer().NetworkFlow()
	src, dst := netFlow.Endpoints()
	reverseFlow := gopacket.NewFlow(dst, src)

端点和流都可以被用作 map 键，
并且等号操作符可以比较它们，
因此您可以通过对端点的判断条件轻松地将所有数据包分组在一起：

	flows := map[gopacket.Endpoint]chan gopacket.Packet
	packet := gopacket.NewPacket(myPacketData, layers.LayerTypeEthernet, gopacket.Lazy)
	// 根据TCP数据包的目标端口，将所有数据吧发送到不同的通道。
	if tcp := packet.Layer(layers.LayerTypeTCP); tcp != nil {
	  flows[tcp.TransportFlow().Dst()] <- packet
	}
	// 寻找有着相同源和目标网络地址的数据包
	if net := packet.NetworkLayer(); net != nil {
	  src, dst := net.NetworkFlow().Endpoints()
	  if src == dst {
	    fmt.Println("有着相同源和目标地址的可疑数据包: %s", src)
	  }
	}
	// 寻找所有从UDP端口1000到UDP端口500的数据包
	interestingFlow := gopacket.FlowFromEndpoints(layers.NewUDPPortEndpoint(1000), layers.NewUDPPortEndpoint(500))
	if t := packet.NetworkLayer(); t != nil && t.TransportFlow() == interestingFlow {
	  fmt.Println("发现了我所寻找的UDP数据流！")
	}

为了平衡负载，Flow 和 Endpoint 都有 FastHash() 函数，它们提供了快速的、非加密的哈希值。
特别重要的是 Flow FastHash() 是对称运算：A->B 的哈希值与 B->A 的哈希值相同。
一个可能的使用示例：

	channels := [8]chan gopacket.Packet
	for i := 0; i < 8; i++ {
	  channels[i] = make(chan gopacket.Packet)
	  go packetHandler(channels[i])
	}
	for packet := range getPackets() {
	  if net := packet.NetworkLayer(); net != nil {
	    channels[int(net.NetworkFlow().FastHash()) & 0x7] <- packet
	  }
	}

这允许我们在确保每个传输流将所有数据包看作同一个数据流
（从数据流的另一端也没问题）的情况下，对数据包进行分割。


实现你自己的解码器


如果你的网络有一些奇怪的封装，你可以实现自己的解码器。
在这个例子中，我们处理以太网数据包，但这些数据包被封装在一个 4 字节的头部中。

	// 创建一个层类型，并给它一个唯一的编号，以避免冲突。
	// 给该层定义对应的名字和解码器。
	var MyLayerType = gopacket.RegisterLayerType(12345, gopacket.LayerTypeMetadata{Name: "MyLayerType", Decoder: gopacket.DecodeFunc(decodeMyLayer)})

	// 实现 MyLayer
	type MyLayer struct {
	  StrangeHeader []byte
	  payload []byte
	}
	func (m MyLayer) LayerType() gopacket.LayerType { return MyLayerType }
	func (m MyLayer) LayerContents() []byte { return m.StrangeHeader }
	func (m MyLayer) LayerPayload() []byte { return m.payload }

	// 现在实现一个解码器... 这个解码器会去掉数据包的前 4 个字节。
	func decodeMyLayer(data []byte, p gopacket.PacketBuilder) error {
	  // 添加 MyLayer
	  p.AddLayer(&MyLayer{data[:4], data[4:]})
	  // 决定如何处理数据包的剩余部分
	  return p.NextDecoder(layers.LayerTypeEthernet)
	}

	// 最后，解码你的数据包：
	p := gopacket.NewPacket(data, MyLayerType, gopacket.Lazy)

See the docs for Decoder and PacketBuilder for more details on how coding
decoders works, or look at RegisterLayerType and RegisterEndpointType to see how
to add layer/endpoint types to gopacket.
您可以去查看 Decoder 和 PacketBuilder 的文档，以获取更多关于编写解码器的细节，
或者查看 RegisterLayerType 和 RegisterEndpointType 以了解如何向 gopacket 添加层/端点类型。

# 使用 DecodingLayerParser 快速解码

你的文档写太长了，我本来是不会看的，
但你的总结又很好地弥补了这一点（Too Long; Didn't Read. TLDR）：
DecodingLayerParser 解码数据包只需要 NewPacket 解码的 10% 时间，
但它仅支持具有已知协议栈的数据包。


使用 gopacket.NewPacket 或 PacketSource.Packets 的基本解码会有一些慢，
这是因为它需要为每个数据包和每个相应的层分配新的内存。
它非常灵活，可以处理所有已知的层类型，但有时您只关心特定的一组层，
因此有时这种灵活性是多余的。

DecodingLayerParser 通过直接将数据包层解码到预分配的对象中来避免内存分配，
然后您可以引用这些对象来获取数据包的信息。
一个快速示例如下：

	func main() {
	  var eth layers.Ethernet
	  var ip4 layers.IPv4
	  var ip6 layers.IPv6
	  var tcp layers.TCP
	  parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &eth, &ip4, &ip6, &tcp)
	  decoded := []gopacket.LayerType{}
	  for packetData := range somehowGetPacketData() {
	    if err := parser.DecodeLayers(packetData, &decoded); err != nil {
	      fmt.Fprintf(os.Stderr, "无法解码网络层: %v\n", err)
	      continue
	    }
	    for _, layerType := range decoded {
	      switch layerType {
	        case layers.LayerTypeIPv6:
	          fmt.Println("    IP6 ", ip6.SrcIP, ip6.DstIP)
	        case layers.LayerTypeIPv4:
	          fmt.Println("    IP4 ", ip4.SrcIP, ip4.DstIP)
	      }
	    }
	  }
	}

值得注意的是，解析器（Parser）修改了传入的层（eth、ip4、ip6、tcp）而不是分配新的层，
从而大大加快了解码过程。它甚至根据层类型进行分支……
它将只会处理（eth、ip4、tcp）或（eth、ip6、tcp）协议栈。而不会处理任何其他类型……
因为没有传入其他解码器，所以（eth、ip4、udp）协议栈将会在 ip4 后停止解码，
并且只通过“decoded”切片返回 [LayerTypeEthernet, LayerTypeIPv4]（以及一个error，表示无法解码 UDP 数据包）。

很遗憾，并非所有层都可以被 DecodingLayerParser 使用……
只有实现了 DecodingLayer 接口的层才能被使用。
此外，是有可能创建 DecodingLayer 的，而这些 DecodingLayer 本身并不是 Layer……
请查看 layers.IPv6ExtensionSkipper 以获取这方面的示例。

使用 DecodingLayerContainer 来快速并自定义解码

尽管它很灵活，但在某些情况下，这种解决方案可能不是最佳的。
例如，如果您只有几个层，那么稀疏数组索引或线性数组扫描可能提供更快的操作。

为了适应这些情况，gopacket引入了 DecodingLayerContainer 接口及其实现：
DecodingLayerSparse、DecodingLayerArray 和 DecodingLayerMap。
您可以使用 SetDecodingLayerContainer 方法为 DecodingLayerParser 指定容器实现。
示例如下：

	dlp := gopacket.NewDecodingLayerParser(LayerTypeEthernet)
	dlp.SetDecodingLayerContainer(gopacket.DecodingLayerSparse(nil))
	var eth layers.Ethernet
	dlp.AddDecodingLayer(&eth)
	// ... 如同正常情况一样添加层并且使用 DecodingLayerParser...

To skip one level of indirection (though sacrificing some capabilities) you may
also use DecodingLayerContainer as a decoding tool as it is. In this case you have to
handle unknown layer types and layer panics by yourself. Example:
若想要跳过一层"套娃"（Level of Indirection) (尽管这会牺牲一些功能），
您也可以将 DecodingLayerContainer 作为解码工具使用。

	func main() {
	  var eth layers.Ethernet
	  var ip4 layers.IPv4
	  var ip6 layers.IPv6
	  var tcp layers.TCP
	  dlc := gopacket.DecodingLayerContainer(gopacket.DecodingLayerArray(nil))
	  dlc = dlc.Put(&eth)
	  dlc = dlc.Put(&ip4)
	  dlc = dlc.Put(&ip6)
	  dlc = dlc.Put(&tcp)
	  // 你也可以指定一些有意义的 DecodeFeedback
	  decoder := dlc.LayersDecoder(LayerTypeEthernet, gopacket.NilDecodeFeedback)
	  decoded := make([]gopacket.LayerType, 0, 20)
	  for packetData := range somehowGetPacketData() {
	    lt, err := decoder(packetData, &decoded)
	    if err != nil {
	      fmt.Fprintf(os.Stderr, "解码网络层失败: %v\n", err)
	      continue
	    }
	    if lt != gopacket.LayerTypeZero {
	      fmt.Fprintf(os.Stderr, "未知的层类型: %v\n", lt)
	      continue
	    }
	    for _, layerType := range decoded {
		  // 像上面展示的那样检查解码的层类型
	    }
	  }
	}

当层（layers）所使用的 LayerType 值可以被解码，
并且层的数量不多时，DecodingLayerSparse 是最快速及最高效的，
当层的数量很多时，DecodingLayerSparse 可能会导致更大的内存占用。
DecodingLayerArray 则十分紧凑，并且主要用于解码层数量不多的情况
（最多 ~10-15，但请自行进行基准测试），
而 DecodingLayerMap 则更加灵活，同时也是 DecodingLayerParser 的默认选择。
请参阅 layers 子包中的测试和基准测试，以进一步查看使用示例和性能测量。

如果您想使用自己的内部数据包解码逻辑，
也可以选择实现自己的 DecodingLayerContainer。

创建数据包

除了提供解码数据包的能力，gopacket 还允许您从头开始创建数据包。
许多 gopacket 层实现了 SerializableLayer 接口；
这些层可以按以下方式序列化为 []byte 数组：


	ip := &layers.IPv4{
	  SrcIP: net.IP{1, 2, 3, 4},
	  DstIP: net.IP{5, 6, 7, 8},
	  // 诸如此类...
	}
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{}  // 您可以从 SerializeOptions 中了解到具体细节.
	err := ip.SerializeTo(buf, opts)
	if err != nil { panic(err) }
	fmt.Println(buf.Bytes())  // 输出一个包含序列化 IPv4 层的字节切片。

SerializeTO 将给定的层添加到 SerializeBuffer 的**首部**，
并将当前缓冲区的 Bytes() 切片视为当前序列化层的有效载荷。
因此，您可以通过以相反的顺序序列化一组层（例如，Payload、TCP、IP、Ethernet）
来序列化整个数据包。
SerializeBuffer 的 SerializeLayers 函数是一个可以做到这一点的辅助函数。

生成一个（空的且无用的，因为没有设置任何字段）以太网数据包，
例如，您可以运行：

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{}
	gopacket.SerializeLayers(buf, opts,
	  &layers.Ethernet{},
	  &layers.IPv4{},
	  &layers.TCP{},
	  gopacket.Payload([]byte{1, 2, 3, 4}))
	packetData := buf.Bytes()

最后一点提醒

如果您使用 gopacket，您几乎肯定会希望确保导入了 gopacket/layers，
因为当导入时，它会设置所有的 LayerType 变量，
并填充了许多有趣的变量/映射（DecodersByLayerName 等）。
因此，即使您不直接使用任何 layers 函数，也建议您使用以下导入：

	import (
	  _ "github.com/tochusc/gopacket/layers"
	)

	TochusC AOSP Lab. 2024

*/

package gopacket
