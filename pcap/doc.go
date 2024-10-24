// Copyright 2024 TochusC. All rights reserved.
// Copyright 2012 Google, Inc. All rights reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file in the root of the source
// tree.

/*
pcap 允许 gopacket 的使用者从网络或者 pcap 文件中读取数据包。

pcap 包是 gopacket 的子包，http://github.com/tochusc/gopacket，
虽然它也可以独立使用，如果你仅仅只是想从网络中获取数据包的话。

Depending on libpcap version, os support, or file timestamp resolution,
nanosecond resolution is used for the internal timestamps. Returned timestamps
are always scaled to nanosecond resolution due to the usage of time.Time.
libpcap must be at least version 1.5 to support nanosecond timestamps. OpenLive
supports only microsecond resolution.
依赖于 libpcap 的版本，操作系统的支持，或文件时间戳的分辨率的不同，
pcap 内部时间戳使用的是纳秒分辨率。
因为使用了 time.Time，pcap 返回的时间戳总是缩放到纳秒分辨率。
libpcap 必须至少是 1.5 版本才支持纳秒时间戳，而OpenLive 则仅支持微秒分辨率。

读取 PCAP 文件

以下代码可以用来从一个 pcap 文件中读取数据包。

	if handle, err := pcap.OpenOffline("/path/to/my/file"); err != nil {
	  panic(err)
	} else {
	  packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	  for packet := range packetSource.Packets() {
	    handlePacket(packet)  // 使用数据包做些什么事情之类的
	  }
	}

读取网络数据包

以下代码可以用来从一个网络设备中读取数据包，比如 "eth0"。
注意，OpenLive 仅支持微秒分辨率。

	if handle, err := pcap.OpenLive("eth0", 1600, true, pcap.BlockForever); err != nil {
	  panic(err)
	} else if err := handle.SetBPFFilter("tcp and port 80"); err != nil {  // optional
	  panic(err)
	} else {
	  packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	  for packet := range packetSource.Packets() {
	    handlePacket(packet)  // 使用数据包做些什么事情之类的
	  }
	}

未激活句柄

较新版本的 PCAP 中引入了“未激活”PCAP句柄的概念。
现在，用户不再需要不断地向 pcap_open_live 添加新的参数，而是调用 pcap_create 来创建一个句柄，
然后使用一堆可选的函数调用来设置它，再调用 pcap_activate 来激活它。
该库也引入了这种机制，来为那些想要体验/使用这些新功能的人提供支持：

	inactive, err := pcap.NewInactiveHandle(deviceName)
	if err != nil {
	  log.Fatal(err)
	}
	defer inactive.CleanUp()

	// 调用 inactive 上的各种函数来设置句柄
	if err = inactive.SetTimeout(time.Minute); err != nil {
	  log.Fatal(err)
	} else if err = inactive.SetTimestampSource("foo"); err != nil {
	  log.Fatal(err)
	}

	// 最后，通过调用 Activate 来创建实际的句柄：
	handle, err := inactive.Activate()  // after this, inactive is no longer valid
	if err != nil {
	  log.Fatal(err)
	}
	defer handle.Close()

	// Now use your handle as you see fit.

# PCAP 超时

pcap.OpenLive 以及 pcap.SetTimeout 都接受超时参数。
如果你不关心超时时间，只需传入 BlockForever，
不用麻烦，这样一般会按照你的预期来运作。

将超时时间设置为 0 不是推荐行为。一些平台，比如 Macs，
(http://www.manpages.info/macosx/pcap.3.html) 写到:

	读取超时时间（Read Timeout）的作用是，当一个数据包被接收时，
	不一定立即返回，而是等待一段时间，来允许更多的数据包到达，
	并且一次从操作系统内核中读取多个数据包。

这意味着如果你只捕获一个数据包，内核可能会决定等待“超时“时间，
以便更多的数据包与之一起返回。因此，超时时间为 0 意味着“永远等待更多的数据包”，
这样...很不好。

为了解决这个问题，pcap 引入了以下行为：
如果传入了负超时时间，我们会在句柄中设置正超时时间，
然后当产生超时错误时，pcap 会在 ReadPacketData/ZeroCopyReadPacketData 中进行内部循环，

# PCAP 文件写入

pcap 包并没有实现 PCAP 文件写入。
但是，gopacket/pcapgo 实现了！ 如果您想写 PCAP 文件，请去查看那里。

# Windows用户的注意事项

gopacket 可以使用 winpcap 或 npcap。如果两者同时安装，npcap 会被优先使用。
确保正确的 Windows 服务已经被加载（npcap 用于 npcap，npf 用于 winpcap）。

TochusC AOSP Lab. 2024
*/
package pcap
