// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package node

import "github.com/atomix/runtime/pkg/logging"

var log = logging.GetLogger()

func New(opts ...Option) *MultiRaftServer {
	var options Options
	options.apply(opts...)
	return &MultiRaftServer{
		Options: options,
	}
}

type MultiRaftServer struct {
	Options
}

func (s *MultiRaftServer) Start() error {

}

func (s *MultiRaftServer) Stop() error {

}
