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

// 用于实现单个 LayerType->DecodingLayer 映射的容器。
type decodingLayerElem struct {
	typ LayerType
	dec DecodingLayer
}

// DecodingLayer 接口表示了能够解码自身的数据包数据层。
//
// DecodingLayer 的重要一点是它们在原地解码自己。
// 在 DecodingLayer 上调用 DecodeFromBytes 会完全重置整个数据层，使其状态由传入的数据定义。
// 返回的错误会使 DecodingLayer 处于未知的中间状态，因此不应信任其中的任何字段。
//
// 由于 DecodingLayer 是在重置自己的字段，因此对 DecodeFromBytes 的调用通常不需要任何内存分配。
type DecodingLayer interface {
	// DecodeFromBytes 重置这个数据层的内部状态，使其状态由传入的字节定义。
	// DecodingLayer 中的切片可能引用传入的数据，
	// 因此如果需要修改传入 DecodingLayer 的原始数据，
	// 请将传入参数设置为原始数据的副本，或确保在修改时 DecodingLayer 已不再使用。
	DecodeFromBytes(data []byte, df DecodeFeedback) error
	// CanDecode 返回这个 DecodingLayer 可以解码的 LayerTypes 集合。
	// 对于同时也是 DecodingLayers 的 Layer，这通常是该 Layer 的 LayerType()。
	CanDecode() LayerClass
	// NextLayerType 返回应该用于解码 LayerPayload 的 LayerType。
	NextLayerType() LayerType
	// LayerPayload 是调用 DecodeFromBytes 后剩余的字节集合。
	LayerPayload() []byte
}

// DecodingLayerFunc decodes given packet and stores decoded LayerType
// values into specified slice. Returns either first encountered
// unsupported LayerType value or decoding error. In case of success,
// returns (LayerTypeZero, nil).
// DecodingLayerFUnc 解码给定的数据包，并将解码后的 LayerType 值存储到指定的切片中。
// - 如果成功，则返回 (LayerTypeZero, nil)。
// - 失败则返回解码过程中第一个不支持的 LayerType 值及 解码错误。
type DecodingLayerFunc func([]byte, *[]LayerType) (LayerType, error)

// DecodingLayerContainer 存储所有 DecodingLayer，
// 并作为 DecodingLayerParser 的搜索工具。
type DecodingLayerContainer interface {
	// Put adds new DecodingLayer to container. The new instance of
	// the same DecodingLayerContainer is returned so it may be
	// implemented as a value receiver.
	Put(DecodingLayer) DecodingLayerContainer
	// Decoder returns DecodingLayer to decode given LayerType and
	// true if it was found. If no decoder found, return false.
	Decoder(LayerType) (DecodingLayer, bool)
	// LayersDecoder returns DecodingLayerFunc which decodes given
	// packet, starting with specified LayerType and DecodeFeedback.
	LayersDecoder(first LayerType, df DecodeFeedback) DecodingLayerFunc
}

// DecodingLayerSparse is a sparse array-based implementation of
// DecodingLayerContainer. Each DecodingLayer is addressed in an
// allocated slice by LayerType value itself. Though this is the
// fastest container it may be memory-consuming if used with big
// LayerType values.
type DecodingLayerSparse []DecodingLayer

// Put implements DecodingLayerContainer interface.
func (dl DecodingLayerSparse) Put(d DecodingLayer) DecodingLayerContainer {
	maxLayerType := LayerType(len(dl) - 1)
	for _, typ := range d.CanDecode().LayerTypes() {
		if typ > maxLayerType {
			maxLayerType = typ
		}
	}

	if extra := maxLayerType - LayerType(len(dl)) + 1; extra > 0 {
		dl = append(dl, make([]DecodingLayer, extra)...)
	}

	for _, typ := range d.CanDecode().LayerTypes() {
		dl[typ] = d
	}
	return dl
}

// LayersDecoder implements DecodingLayerContainer interface.
func (dl DecodingLayerSparse) LayersDecoder(first LayerType, df DecodeFeedback) DecodingLayerFunc {
	return LayersDecoder(dl, first, df)
}

// Decoder implements DecodingLayerContainer interface.
func (dl DecodingLayerSparse) Decoder(typ LayerType) (DecodingLayer, bool) {
	if int64(typ) < int64(len(dl)) {
		decoder := dl[typ]
		return decoder, decoder != nil
	}
	return nil, false
}

// DecodingLayerArray is an array-based implementation of
// DecodingLayerContainer. Each DecodingLayer is searched linearly in
// an allocated slice in one-by-one fashion.
type DecodingLayerArray []decodingLayerElem

// Put implements DecodingLayerContainer interface.
func (dl DecodingLayerArray) Put(d DecodingLayer) DecodingLayerContainer {
TYPES:
	for _, typ := range d.CanDecode().LayerTypes() {
		for i := range dl {
			if dl[i].typ == typ {
				dl[i].dec = d
				continue TYPES
			}
		}
		dl = append(dl, decodingLayerElem{typ, d})
	}
	return dl
}

// Decoder implements DecodingLayerContainer interface.
func (dl DecodingLayerArray) Decoder(typ LayerType) (DecodingLayer, bool) {
	for i := range dl {
		if dl[i].typ == typ {
			return dl[i].dec, true
		}
	}
	return nil, false
}

// LayersDecoder implements DecodingLayerContainer interface.
func (dl DecodingLayerArray) LayersDecoder(first LayerType, df DecodeFeedback) DecodingLayerFunc {
	return LayersDecoder(dl, first, df)
}

// DecodingLayerMap is an map-based implementation of
// DecodingLayerContainer. Each DecodingLayer is searched in a map
// hashed by LayerType value.
type DecodingLayerMap map[LayerType]DecodingLayer

// Put implements DecodingLayerContainer interface.
func (dl DecodingLayerMap) Put(d DecodingLayer) DecodingLayerContainer {
	for _, typ := range d.CanDecode().LayerTypes() {
		if dl == nil {
			dl = make(map[LayerType]DecodingLayer)
		}
		dl[typ] = d
	}
	return dl
}

// Decoder implements DecodingLayerContainer interface.
func (dl DecodingLayerMap) Decoder(typ LayerType) (DecodingLayer, bool) {
	d, ok := dl[typ]
	return d, ok
}

// LayersDecoder implements DecodingLayerContainer interface.
func (dl DecodingLayerMap) LayersDecoder(first LayerType, df DecodeFeedback) DecodingLayerFunc {
	return LayersDecoder(dl, first, df)
}

// Static code check.
var (
	_ = []DecodingLayerContainer{
		DecodingLayerSparse(nil),
		DecodingLayerMap(nil),
		DecodingLayerArray(nil),
	}
)

// DecodingLayerParser parses a given set of layer types.  See DecodeLayers for
// more information on how DecodingLayerParser should be used.
type DecodingLayerParser struct {
	// DecodingLayerParserOptions is the set of options available to the
	// user to define the parser's behavior.
	DecodingLayerParserOptions
	dlc   DecodingLayerContainer
	first LayerType
	df    DecodeFeedback

	decodeFunc DecodingLayerFunc

	// Truncated is set when a decode layer detects that the packet has been
	// truncated.
	Truncated bool
}

// AddDecodingLayer adds a decoding layer to the parser.  This adds support for
// the decoding layer's CanDecode layers to the parser... should they be
// encountered, they'll be parsed.
func (l *DecodingLayerParser) AddDecodingLayer(d DecodingLayer) {
	l.SetDecodingLayerContainer(l.dlc.Put(d))
}

// SetTruncated is used by DecodingLayers to set the Truncated boolean in the
// DecodingLayerParser.  Users should simply read Truncated after calling
// DecodeLayers.
func (l *DecodingLayerParser) SetTruncated() {
	l.Truncated = true
}

// NewDecodingLayerParser creates a new DecodingLayerParser and adds in all
// of the given DecodingLayers with AddDecodingLayer.
//
// Each call to DecodeLayers will attempt to decode the given bytes first by
// treating them as a 'first'-type layer, then by using NextLayerType on
// subsequently decoded layers to find the next relevant decoder.  Should a
// deoder not be available for the layer type returned by NextLayerType,
// decoding will stop.
//
// NewDecodingLayerParser uses DecodingLayerMap container by
// default.
func NewDecodingLayerParser(first LayerType, decoders ...DecodingLayer) *DecodingLayerParser {
	dlp := &DecodingLayerParser{first: first}
	dlp.df = dlp // Cast this once to the interface
	// default container
	dlc := DecodingLayerContainer(DecodingLayerMap(make(map[LayerType]DecodingLayer)))
	for _, d := range decoders {
		dlc = dlc.Put(d)
	}

	dlp.SetDecodingLayerContainer(dlc)
	return dlp
}

// SetDecodingLayerContainer specifies container with decoders. This
// call replaces all decoders already registered in given instance of
// DecodingLayerParser.
func (l *DecodingLayerParser) SetDecodingLayerContainer(dlc DecodingLayerContainer) {
	l.dlc = dlc
	l.decodeFunc = l.dlc.LayersDecoder(l.first, l.df)
}

// DecodeLayers decodes as many layers as possible from the given data.  It
// initially treats the data as layer type 'typ', then uses NextLayerType on
// each subsequent decoded layer until it gets to a layer type it doesn't know
// how to parse.
//
// For each layer successfully decoded, DecodeLayers appends the layer type to
// the decoded slice.  DecodeLayers truncates the 'decoded' slice initially, so
// there's no need to empty it yourself.
//
// This decoding method is about an order of magnitude faster than packet
// decoding, because it only decodes known layers that have already been
// allocated.  This means it doesn't need to allocate each layer it returns...
// instead it overwrites the layers that already exist.
//
// Example usage:
//
//	func main() {
//	  var eth layers.Ethernet
//	  var ip4 layers.IPv4
//	  var ip6 layers.IPv6
//	  var tcp layers.TCP
//	  var udp layers.UDP
//	  var payload gopacket.Payload
//	  parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &eth, &ip4, &ip6, &tcp, &udp, &payload)
//	  var source gopacket.PacketDataSource = getMyDataSource()
//	  decodedLayers := make([]gopacket.LayerType, 0, 10)
//	  for {
//	    data, _, err := source.ReadPacketData()
//	    if err != nil {
//	      fmt.Println("Error reading packet data: ", err)
//	      continue
//	    }
//	    fmt.Println("Decoding packet")
//	    err = parser.DecodeLayers(data, &decodedLayers)
//	    for _, typ := range decodedLayers {
//	      fmt.Println("  Successfully decoded layer type", typ)
//	      switch typ {
//	        case layers.LayerTypeEthernet:
//	          fmt.Println("    Eth ", eth.SrcMAC, eth.DstMAC)
//	        case layers.LayerTypeIPv4:
//	          fmt.Println("    IP4 ", ip4.SrcIP, ip4.DstIP)
//	        case layers.LayerTypeIPv6:
//	          fmt.Println("    IP6 ", ip6.SrcIP, ip6.DstIP)
//	        case layers.LayerTypeTCP:
//	          fmt.Println("    TCP ", tcp.SrcPort, tcp.DstPort)
//	        case layers.LayerTypeUDP:
//	          fmt.Println("    UDP ", udp.SrcPort, udp.DstPort)
//	      }
//	    }
//	    if decodedLayers.Truncated {
//	      fmt.Println("  Packet has been truncated")
//	    }
//	    if err != nil {
//	      fmt.Println("  Error encountered:", err)
//	    }
//	  }
//	}
//
// If DecodeLayers is unable to decode the next layer type, it will return the
// error UnsupportedLayerType.
func (l *DecodingLayerParser) DecodeLayers(data []byte, decoded *[]LayerType) (err error) {
	l.Truncated = false
	if !l.IgnorePanic {
		defer panicToError(&err)
	}
	typ, err := l.decodeFunc(data, decoded)
	if typ != LayerTypeZero {
		// no decoder
		if l.IgnoreUnsupported {
			return nil
		}
		return UnsupportedLayerType(typ)
	}
	return err
}

// UnsupportedLayerType is returned by DecodingLayerParser if DecodeLayers
// encounters a layer type that the DecodingLayerParser has no decoder for.
type UnsupportedLayerType LayerType

// Error implements the error interface, returning a string to say that the
// given layer type is unsupported.
func (e UnsupportedLayerType) Error() string {
	return fmt.Sprintf("No decoder for layer type %v", LayerType(e))
}

func panicToError(e *error) {
	if r := recover(); r != nil {
		*e = fmt.Errorf("panic: %v", r)
	}
}

// DecodingLayerParserOptions provides options to affect the behavior of a given
// DecodingLayerParser.
type DecodingLayerParserOptions struct {
	// IgnorePanic determines whether a DecodingLayerParser should stop
	// panics on its own (by returning them as an error from DecodeLayers)
	// or should allow them to raise up the stack.  Handling errors does add
	// latency to the process of decoding layers, but is much safer for
	// callers.  IgnorePanic defaults to false, thus if the caller does
	// nothing decode panics will be returned as errors.
	IgnorePanic bool
	// IgnoreUnsupported will stop parsing and return a nil error when it
	// encounters a layer it doesn't have a parser for, instead of returning an
	// UnsupportedLayerType error.  If this is true, it's up to the caller to make
	// sure that all expected layers have been parsed (by checking the decoded
	// slice).
	IgnoreUnsupported bool
}
