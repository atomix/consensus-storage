// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package node

import (
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/primitive"
)

const (
	defaultAPIPort  = 8080
	defaultRaftPort = 5000
	defaultDataDir  = "/var/lib/atomix/data"
)

type Options struct {
	NodeID           multiraftv1.NodeID
	Config           multiraftv1.MultiRaftConfig
	PrimitiveService PrimitiveServiceOptions
	RaftService      RaftServiceOptions
	PrimitiveTypes   []primitive.Type
}

func (o *Options) apply(opts ...Option) {
	o.PrimitiveService.Port = defaultAPIPort
	o.RaftService.Port = defaultRaftPort
	for _, opt := range opts {
		opt(o)
	}
}

type Option func(*Options)

type ServiceOptions struct {
	Host string
	Port int
}

type PrimitiveServiceOptions struct {
	ServiceOptions
}

type RaftServiceOptions struct {
	ServiceOptions
}

func WithOptions(opts Options) Option {
	return func(options *Options) {
		*options = opts
	}
}

func WithNodeID(nodeID multiraftv1.NodeID) Option {
	return func(options *Options) {
		options.NodeID = nodeID
	}
}

func WithConfig(config multiraftv1.MultiRaftConfig) Option {
	return func(options *Options) {
		options.Config = config
	}
}

func WithPrimitiveTypes(primitiveTypes ...primitive.Type) Option {
	return func(options *Options) {
		options.PrimitiveTypes = append(options.PrimitiveTypes, primitiveTypes...)
	}
}

func WithPrimitiveHost(host string) Option {
	return func(options *Options) {
		options.PrimitiveService.Host = host
	}
}

func WithPrimitivePort(port int) Option {
	return func(options *Options) {
		options.PrimitiveService.Port = port
	}
}

func WithRaftHost(host string) Option {
	return func(options *Options) {
		options.RaftService.Host = host
	}
}

func WithRaftPort(port int) Option {
	return func(options *Options) {
		options.RaftService.Port = port
	}
}
