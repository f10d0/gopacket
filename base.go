// Copyright 2024 TochusC AOSP Lab. All rights reserved.
// Copyright 2012 Google, Inc. All rights reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file in the root of the source
// tree.

package gopacket

import (
	"fmt"
)

// 在解码时，数据包的数据被分解为多个数据层。
// 调用者可以调用 LayerType() 来确定他们从数据包中收到的数据层类型。
// 另外，也可以使用类型断言来获取实际的数据层类型，以深入检查数据。
type Layer interface {
	// LayerType 是这个协议层的 gopacket 类型。
	LayerType() LayerType
	// LayerContents 返回构成这个协议层的字节集合。
	LayerContents() []byte
	// LayerPayload 返回这个层所承载的字节集合（即有效载荷），不包括这个层本身。
	LayerPayload() []byte
}

// Payload 是一个包含数据包有效载荷的数据层。
// 有效载荷的定义取决于之前的数据层；对于 TCP 和 UDP，我们在第 4 层以上停止解码，
// 并将剩余的字节作为 Payload 返回，此时返回的 Payload 是一个 ApplicationLayer。
type Payload []byte

// LayerType 返回 LayerTypePayload
func (p Payload) LayerType() LayerType { return LayerTypePayload }

// LayerContents 返回构成 Payload 层的字节集合。
func (p Payload) LayerContents() []byte { return []byte(p) }

// LayerPayload 返回 Payload 层所承载的字节集合（即有效载荷），不包括这个层本身。
func (p Payload) LayerPayload() []byte { return nil }

// Payload 以字节形式返回 Payload 层。
func (p Payload) Payload() []byte { return []byte(p) }

// String 实现 fmt.Stringer 接口。
func (p Payload) String() string { return fmt.Sprintf("%d byte(s)", len(p)) }

// GoString 实现 fmt.GoStringer 接口。
func (p Payload) GoString() string { return LongBytesGoString([]byte(p)) }

// CanDecode 实现 DecodingLayer 接口。
func (p Payload) CanDecode() LayerClass { return LayerTypePayload }

// NextLayerType 实现 DecodingLayer 接口。
func (p Payload) NextLayerType() LayerType { return LayerTypeZero }

// DecodeFromBytes 实现 DecodingLayer 接口。
func (p *Payload) DecodeFromBytes(data []byte, df DecodeFeedback) error {
	*p = Payload(data)
	return nil
}

// SerializeTo 将这个层的序列化形式写入 SerializationBuffer 中，
// 实现了 gopacket.SerializableLayer 接口。
// 有关更多信息，请查阅 gopacket.SerializableLayer 的文档。
func (p Payload) SerializeTo(b SerializeBuffer, opts SerializeOptions) error {
	bytes, err := b.PrependBytes(len(p))
	if err != nil {
		return err
	}
	copy(bytes, p)
	return nil
}

// decodePayload 将数据全部解码至一个 Payload 层中。
func decodePayload(data []byte, p PacketBuilder) error {
	payload := &Payload{}
	if err := payload.DecodeFromBytes(data, p); err != nil {
		return err
	}
	p.AddLayer(payload)
	p.SetApplicationLayer(payload)
	return nil
}

// Fragment 是一个包含较大帧的片段的数据层，
// 被 IPv4 和 IPv6 等允许对有效载荷进行分片的层使用。
type Fragment []byte

// LayerType returns LayerTypeFragment
func (p *Fragment) LayerType() LayerType { return LayerTypeFragment }

// LayerContents implements Layer.
func (p *Fragment) LayerContents() []byte { return []byte(*p) }

// LayerPayload implements Layer.
func (p *Fragment) LayerPayload() []byte { return nil }

// Payload returns this layer as a byte slice.
func (p *Fragment) Payload() []byte { return []byte(*p) }

// String implements fmt.Stringer.
func (p *Fragment) String() string { return fmt.Sprintf("%d byte(s)", len(*p)) }

// CanDecode implements DecodingLayer.
func (p *Fragment) CanDecode() LayerClass { return LayerTypeFragment }

// NextLayerType implements DecodingLayer.
func (p *Fragment) NextLayerType() LayerType { return LayerTypeZero }

// DecodeFromBytes implements DecodingLayer.
func (p *Fragment) DecodeFromBytes(data []byte, df DecodeFeedback) error {
	*p = Fragment(data)
	return nil
}

// SerializeTo 将这个层的序列化形式写入 SerializationBuffer 中，
// 其实现了 gopacket.SerializableLayer 接口。
// 有关更多信息，请查阅 gopacket.SerializableLayer 的文档。
func (p *Fragment) SerializeTo(b SerializeBuffer, opts SerializeOptions) error {
	bytes, err := b.PrependBytes(len(*p))
	if err != nil {
		return err
	}
	copy(bytes, *p)
	return nil
}

// decodeFragment decodes data by returning it all in a Fragment layer.
func decodeFragment(data []byte, p PacketBuilder) error {
	payload := &Fragment{}
	if err := payload.DecodeFromBytes(data, p); err != nil {
		return err
	}
	p.AddLayer(payload)
	p.SetApplicationLayer(payload)
	return nil
}

// These layers correspond to Internet Protocol Suite (TCP/IP) layers, and their
// corresponding OSI layers, as best as possible.

// LinkLayer is the packet layer corresponding to TCP/IP layer 1 (OSI layer 2)
type LinkLayer interface {
	Layer
	LinkFlow() Flow
}

// NetworkLayer is the packet layer corresponding to TCP/IP layer 2 (OSI
// layer 3)
type NetworkLayer interface {
	Layer
	NetworkFlow() Flow
}

// TransportLayer is the packet layer corresponding to the TCP/IP layer 3 (OSI
// layer 4)
type TransportLayer interface {
	Layer
	TransportFlow() Flow
}

// ApplicationLayer is the packet layer corresponding to the TCP/IP layer 4 (OSI
// layer 7), also known as the packet payload.
type ApplicationLayer interface {
	Layer
	Payload() []byte
}

// ErrorLayer is a packet layer created when decoding of the packet has failed.
// Its payload is all the bytes that we were unable to decode, and the returned
// error details why the decoding failed.
type ErrorLayer interface {
	Layer
	Error() error
}
