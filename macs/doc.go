// Copyright 2024 TochusC AOSP Lab. All rights reserved.
// Copyright 2012 Google, Inc. All rights reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file in the root of the source
// tree.

// macs 包提供了所有有效以太网 MAC 地址前缀到其关联组织的内存映射。
//
// ValidMACPrefixMap 映射了 3 字节前缀到组织字符串的映射。
// 并且可以使用 'go run gen.go' 来更新。
package macs
