// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package multiraft

func NewNode(opts ...Option) *Node {
	var options Options
	options.apply(opts...)
	return &Node{
		Options: options,
		...
	}
}

type Node struct {
	Options
}
