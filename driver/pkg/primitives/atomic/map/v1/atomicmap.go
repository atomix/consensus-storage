// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	api "github.com/atomix/multi-raft-storage/api/atomix/multiraft/atomic/map/v1"
	"github.com/atomix/multi-raft-storage/driver/pkg/client"
	atomicmapv1 "github.com/atomix/runtime/api/atomix/runtime/atomic/map/v1"
	"github.com/atomix/runtime/sdk/pkg/logging"
)

var log = logging.GetLogger()

func NewAtomicMapServer(protocol *client.Protocol, config api.AtomicMapConfig) atomicmapv1.AtomicMapServer {
	var server = newMultiRaftAtomicMapServer(protocol)
	if config.Cache.Enabled {
		server = newCachingAtomicMapServer(server, config.Cache)
	}
	return server
}
