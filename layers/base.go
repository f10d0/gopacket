// Copyright 2024 TochusC AOSP Lab. All rights reserved.
// Copyright 2012 Google, Inc. All rights reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file in the root of the source
// tree.

package layers

import (
	"github.com/google/gopacket"
)

// BaseLayer 是一个实现了 Layer 接口中 LayerData 和 LayerPayload 函数的便利结构体。
type BaseLayer struct {
	// Contents 是构成这个层的字集合。
	// 例如：对于一个以太网数据包，这将是构成以太网帧的字节集。
	Contents []byte

	// Payload 是这个层所承载的（但不是这个层的一部分的）字节集合。
	// 再以以太网为例，Payload 的内容 将是以太网协议所封装的字节集合。
	Payload []byte
}

// LayerContents 返回该数据包协议层的字节。
func (b *BaseLayer) LayerContents() []byte { return b.Contents }

// LayerPayload 返回该数据包协议层所承载的字节。
func (b *BaseLayer) LayerPayload() []byte { return b.Payload }

type layerDecodingLayer interface {
	gopacket.Layer
	DecodeFromBytes([]byte, gopacket.DecodeFeedback) error
	NextLayerType() gopacket.LayerType
}

func decodingLayerDecoder(d layerDecodingLayer, data []byte, p gopacket.PacketBuilder) error {
	err := d.DecodeFromBytes(data, p)
	if err != nil {
		return err
	}
	p.AddLayer(d)
	next := d.NextLayerType()
	if next == gopacket.LayerTypeZero {
		return nil
	}
	return p.NextDecoder(next)
}

// 用来清空内存的奇妙方法(Hacky way)...
// 应该会有更好的方法？
// - 译者注：不懂
var lotsOfZeros [1024]byte
