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

// SerializableLayer 接口的实现允许将其写入为一组字节，
// 以便这些字节可以通过网络发送或由调用者使用。
// SerializableLayer 由某些 Layer 类型实现，并可以使用 LayerWriter 对象将其编码为字节。
type SerializableLayer interface {
	// SerializeTo 将该数据层写入到一个字节切片中，
	// 如果需要，字节切片的尺寸将会增大，以适应该数据层中的数据。
	// 参数：
	//   b： 用于写入该数据层的 SerializeBuffer。当调用时，b.Bytes() 返回的是该数据层应该包装的有效载荷。
	//     请注意，数据层可能需要在有效载荷之前（常见）、之后（不常见）或
	//	   两者（有时需要在数据包数据的末尾添加填充或页脚）上进行修改。
	//     同时也可能（尽管可能非常少见）需要覆盖当前有效载荷中的部分字节。
	//     在此调用后 b.Bytes() 应该返回该数据层编码后的结果。
	//   opts：在写出数据时使用的选项。
	// 返回：
	//  如果在编码过程中遇到问题，则返回错误。
	// 	如果返回错误，则应该认为数据中的字节无效，不应使用。
	//
	//  SerializeTo 调用应该完全忽略 LayerContents 和 LayerPayload。
	//  它只是基于结构字段进行序列化，既不修改也不使用内容/有效载荷。
	SerializeTo(b SerializeBuffer, opts SerializeOptions) error
	// LayerType 返回正在写入到缓冲区的数据层的类型
	LayerType() LayerType
}

// SerializeOptions 提供了 SerializableLayer 可能想要实现的行为选项。
type SerializeOptions struct {
	// FixLengths 决定在序列化期间，数据层是否应该修复依赖于 有效载荷 的任何 长度字段 的值。
	FixLengths bool
	// ComputeChecksums 决定在序列化期间，数据层是否应该基于有效载荷重新计算校验和。
	ComputeChecksums bool
}

// SerializeBuffer 是 gopacket 用于写出数据包数据层的辅助工具。
// SerializeBuffer 从一个空的 []byte 开始。
// 后续的 PrependBytes 调用将返回当前 Bytes() 之前的字节切片，
// AppendBytes 返回当前 Bytes() 之后的字节切片。
//
// PrependBytes/AppendBytes 返回的字节切片不会被清零，
// 所以如果你想确保它们全是零，请自行设置为零。
//
// SerializeBuffer 是专门为处理数据包写入而设计的，
// 数据包写入与普通写入不同，首先从最内层开始写入，然后再逐步向外写入，往往更加容易。
// 这意味着数据包写入经常需要在开头添加字节。
// 而这与使用 append() 写入字节切片的典型写入相悖，
// 在典型写入情况下，我们通常往往只在缓冲区的末尾写入。
//
// 它可以通过 Clear 方法重用。
// 但是请注意，Clear 调用将使任何先前 Bytes() 调用返回的字节切片无效（内置数组被重用）。
//
// 1. 重用写入缓冲区通常比创建新缓冲区快得多，并且在默认实现中，这样避免了额外的内存分配。
// 2. 如果先前 Bytes() 调用返回的字节切片后续仍需使用，最好是创建一个新的 SerializeBuffer。
//
// Clear 方法专门设计用于最小化 SerializeBuffer 上的后续内存分配。
// 例如：如果你进行了一组 Prepend/Append 调用，然后清除，然后使用相同的大小进行相同的调用，
// 第二轮（以及所有类似的后续轮次）不应分配任何新内存。
type SerializeBuffer interface {
	// Bytes 返回到目前为止由 Prepend/Append 调用收集的连续字节集。
	// Bytes 返回的切片将被后续的 Clear 调用修改，所以如果你打算清除这个 SerializeBuffer，
	// 你可能会想先将 Bytes 复制到一个安全的地方。
	Bytes() []byte
	// PrependBytes 返回前置到当前缓冲区的字节切片。
	// 这些字节的起始状态是不确定的，所以调用者应该覆盖它们。
	// 只有在调用者知道他们将立即覆盖所有返回的字节时，才能调用 PrependBytes。
	PrependBytes(num int) ([]byte, error)
	// AppendBytes 返回后置到当前缓冲区的字节切片。
	// 这些字节的起始状态是不确定的，所以调用者应该覆盖它们。
	// 只有在调用者知道他们将立即覆盖所有返回的字节时，才能调用 AppendBytes。
	AppendBytes(num int) ([]byte, error)
	// Clear 重置 SerializeBuffer 为一个新的、空的缓冲区。
	// 在调用 Clear 后，对于该缓冲区的任何先前调用 Bytes() 返回的字节切片都应该被视为无效。
	Clear() error
	// Layers 返回已经成功序列化到这个缓冲区的所有数据层。
	Layers() []LayerType
	// PushLayer 将当前数据层添加到已经序列化到这个缓冲区的数据层列表中。
	PushLayer(LayerType)
}

type serializeBuffer struct {
	data                []byte
	start               int
	prepended, appended int
	layers              []LayerType
}

// NewSerializeBuffer creates a new instance of the default implementation of
// the SerializeBuffer interface.
func NewSerializeBuffer() SerializeBuffer {
	return &serializeBuffer{}
}

// NewSerializeBufferExpectedSize creates a new buffer for serialization, optimized for an
// expected number of bytes prepended/appended.  This tends to decrease the
// number of memory allocations made by the buffer during writes.
func NewSerializeBufferExpectedSize(expectedPrependLength, expectedAppendLength int) SerializeBuffer {
	return &serializeBuffer{
		data:      make([]byte, expectedPrependLength, expectedPrependLength+expectedAppendLength),
		start:     expectedPrependLength,
		prepended: expectedPrependLength,
		appended:  expectedAppendLength,
	}
}

func (w *serializeBuffer) Bytes() []byte {
	return w.data[w.start:]
}

func (w *serializeBuffer) PrependBytes(num int) ([]byte, error) {
	if num < 0 {
		panic("num < 0")
	}
	if w.start < num {
		toPrepend := w.prepended
		if toPrepend < num {
			toPrepend = num
		}
		w.prepended += toPrepend
		length := cap(w.data) + toPrepend
		newData := make([]byte, length)
		newStart := w.start + toPrepend
		copy(newData[newStart:], w.data[w.start:])
		w.start = newStart
		w.data = newData[:toPrepend+len(w.data)]
	}
	w.start -= num
	return w.data[w.start : w.start+num], nil
}

func (w *serializeBuffer) AppendBytes(num int) ([]byte, error) {
	if num < 0 {
		panic("num < 0")
	}
	initialLength := len(w.data)
	if cap(w.data)-initialLength < num {
		toAppend := w.appended
		if toAppend < num {
			toAppend = num
		}
		w.appended += toAppend
		newData := make([]byte, cap(w.data)+toAppend)
		copy(newData[w.start:], w.data[w.start:])
		w.data = newData[:initialLength]
	}
	// Grow the buffer.  We know it'll be under capacity given above.
	w.data = w.data[:initialLength+num]
	return w.data[initialLength:], nil
}

func (w *serializeBuffer) Clear() error {
	w.start = w.prepended
	w.data = w.data[:w.start]
	w.layers = w.layers[:0]
	return nil
}

func (w *serializeBuffer) Layers() []LayerType {
	return w.layers
}

func (w *serializeBuffer) PushLayer(l LayerType) {
	w.layers = append(w.layers, l)
}

// SerializeLayers clears the given write buffer, then writes all layers into it so
// they correctly wrap each other.  Note that by clearing the buffer, it
// invalidates all slices previously returned by w.Bytes()
//
// Example:
//
//	buf := gopacket.NewSerializeBuffer()
//	opts := gopacket.SerializeOptions{}
//	gopacket.SerializeLayers(buf, opts, a, b, c)
//	firstPayload := buf.Bytes()  // contains byte representation of a(b(c))
//	gopacket.SerializeLayers(buf, opts, d, e, f)
//	secondPayload := buf.Bytes()  // contains byte representation of d(e(f)). firstPayload is now invalidated, since the SerializeLayers call Clears buf.
func SerializeLayers(w SerializeBuffer, opts SerializeOptions, layers ...SerializableLayer) error {
	w.Clear()
	for i := len(layers) - 1; i >= 0; i-- {
		layer := layers[i]
		err := layer.SerializeTo(w, opts)
		if err != nil {
			return err
		}
		w.PushLayer(layer.LayerType())
	}
	return nil
}

// SerializePacket is a convenience function that calls SerializeLayers
// on packet's Layers().
// It returns an error if one of the packet layers is not a SerializableLayer.
func SerializePacket(buf SerializeBuffer, opts SerializeOptions, packet Packet) error {
	sls := []SerializableLayer{}
	for _, layer := range packet.Layers() {
		sl, ok := layer.(SerializableLayer)
		if !ok {
			return fmt.Errorf("layer %s is not serializable", layer.LayerType().String())
		}
		sls = append(sls, sl)
	}
	return SerializeLayers(buf, opts, sls...)
}
