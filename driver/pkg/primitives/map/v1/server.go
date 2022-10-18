// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	api "github.com/atomix/multi-raft-storage/api/atomix/multiraft/map/v1"
	"github.com/atomix/multi-raft-storage/driver/pkg/client"
	mapv1 "github.com/atomix/runtime/api/atomix/runtime/map/v1"
	"github.com/atomix/runtime/sdk/pkg/logging"
	"github.com/atomix/runtime/sdk/pkg/runtime"
)

var log = logging.GetLogger()

func NewMapServer(protocol *client.Protocol, spec runtime.PrimitiveSpec) (mapv1.MapServer, error) {
	var server = newMultiRaftMapServer(protocol, spec)
	var config api.MapConfig
	if err := spec.UnmarshalConfig(&config); err != nil {
		return nil, err
	}
	if config.Cache.Enabled {
		server = newCachingMapServer(server, config.Cache)
	}
	return server, nil
}
