// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package node

import (
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/primitive"
)

const (
	defaultPort = 8080
)

type Options struct {
	ServiceOptions
	Config         multiraftv1.NodeConfig
	PrimitiveTypes []primitive.Type
}

func (o *Options) apply(opts ...Option) {
	o.Port = defaultPort
	for _, opt := range opts {
		opt(o)
	}
}

type Option func(*Options)

type ServiceOptions struct {
	Host string
	Port int
}

func WithOptions(opts Options) Option {
	return func(options *Options) {
		*options = opts
	}
}

func WithHost(host string) Option {
	return func(options *Options) {
		options.Host = host
	}
}

func WithPort(port int) Option {
	return func(options *Options) {
		options.Port = port
	}
}

func WithConfig(config multiraftv1.NodeConfig) Option {
	return func(options *Options) {
		options.Config = config
	}
}

func WithPrimitiveTypes(primitiveTypes ...primitive.Type) Option {
	return func(options *Options) {
		options.PrimitiveTypes = append(options.PrimitiveTypes, primitiveTypes...)
	}
}
