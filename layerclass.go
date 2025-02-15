// Copyright 2024 TochusC AOSP Lab. All rights reserved.
// Copyright 2012 Google, Inc. All rights reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file in the root of the source
// tree.

package gopacket

// LayerClass 是一组 LayerTypes，
// 用于从数据包的多种不同类型数据层中的获取某一特点数据层。
type LayerClass interface {
	// Contains returns true if the given layer type should be considered part
	// of this layer class.
	// 如果给定的数据层类型是该 LayerClass 的一部分，则返回 true。
	Contains(LayerType) bool
	// LayerTypes 返回该 LayerClass 中包含的所有数据层类型。
	// 注意：在 LayerClass 的某些实现中，这可能不是一个快速的操作。
	LayerTypes() []LayerType
}

// Contains implements LayerClass.
func (l LayerType) Contains(a LayerType) bool {
	return l == a
}

// LayerTypes implements LayerClass.
func (l LayerType) LayerTypes() []LayerType {
	return []LayerType{l}
}

// LayerClassSlice implements a LayerClass with a slice.
type LayerClassSlice []bool

// Contains returns true if the given layer type should be considered part
// of this layer class.
func (s LayerClassSlice) Contains(t LayerType) bool {
	return int(t) < len(s) && s[t]
}

// LayerTypes returns all layer types in this LayerClassSlice.
// Because of LayerClassSlice's implementation, this could be quite slow.
func (s LayerClassSlice) LayerTypes() (all []LayerType) {
	for i := 0; i < len(s); i++ {
		if s[i] {
			all = append(all, LayerType(i))
		}
	}
	return
}

// NewLayerClassSlice creates a new LayerClassSlice by creating a slice of
// size max(types) and setting slice[t] to true for each type t.  Note, if
// you implement your own LayerType and give it a high value, this WILL create
// a very large slice.
func NewLayerClassSlice(types []LayerType) LayerClassSlice {
	var max LayerType
	for _, typ := range types {
		if typ > max {
			max = typ
		}
	}
	t := make([]bool, int(max+1))
	for _, typ := range types {
		t[typ] = true
	}
	return t
}

// LayerClassMap implements a LayerClass with a map.
type LayerClassMap map[LayerType]bool

// Contains returns true if the given layer type should be considered part
// of this layer class.
func (m LayerClassMap) Contains(t LayerType) bool {
	return m[t]
}

// LayerTypes returns all layer types in this LayerClassMap.
func (m LayerClassMap) LayerTypes() (all []LayerType) {
	for t := range m {
		all = append(all, t)
	}
	return
}

// NewLayerClassMap creates a LayerClassMap and sets map[t] to true for each
// type in types.
func NewLayerClassMap(types []LayerType) LayerClassMap {
	m := LayerClassMap{}
	for _, typ := range types {
		m[typ] = true
	}
	return m
}

// NewLayerClass creates a LayerClass, attempting to be smart about which type
// it creates based on which types are passed in.
func NewLayerClass(types []LayerType) LayerClass {
	for _, typ := range types {
		if typ > maxLayerType {
			// NewLayerClassSlice could create a very large object, so instead create
			// a map.
			return NewLayerClassMap(types)
		}
	}
	return NewLayerClassSlice(types)
}
