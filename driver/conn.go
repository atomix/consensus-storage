// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"context"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/driver/pkg/client"
	counterv1server "github.com/atomix/multi-raft-storage/driver/pkg/primitives/counter/v1"
	countermapv1server "github.com/atomix/multi-raft-storage/driver/pkg/primitives/countermap/v1"
	electionv1server "github.com/atomix/multi-raft-storage/driver/pkg/primitives/election/v1"
	indexedmapv1server "github.com/atomix/multi-raft-storage/driver/pkg/primitives/indexedmap/v1"
	lockv1server "github.com/atomix/multi-raft-storage/driver/pkg/primitives/lock/v1"
	mapv1server "github.com/atomix/multi-raft-storage/driver/pkg/primitives/map/v1"
	multimapv1server "github.com/atomix/multi-raft-storage/driver/pkg/primitives/multimap/v1"
	setv1server "github.com/atomix/multi-raft-storage/driver/pkg/primitives/set/v1"
	valuev1server "github.com/atomix/multi-raft-storage/driver/pkg/primitives/value/v1"
	counterv1 "github.com/atomix/runtime/api/atomix/runtime/counter/v1"
	countermapv1 "github.com/atomix/runtime/api/atomix/runtime/countermap/v1"
	electionv1 "github.com/atomix/runtime/api/atomix/runtime/election/v1"
	indexedmapv1 "github.com/atomix/runtime/api/atomix/runtime/indexedmap/v1"
	lockv1 "github.com/atomix/runtime/api/atomix/runtime/lock/v1"
	mapv1 "github.com/atomix/runtime/api/atomix/runtime/map/v1"
	multimapv1 "github.com/atomix/runtime/api/atomix/runtime/multimap/v1"
	setv1 "github.com/atomix/runtime/api/atomix/runtime/set/v1"
	valuev1 "github.com/atomix/runtime/api/atomix/runtime/value/v1"
	"github.com/atomix/runtime/sdk/pkg/network"
	"github.com/atomix/runtime/sdk/pkg/runtime"
)

func newConn(network network.Network) *multiRaftConn {
	return &multiRaftConn{
		Client: client.NewClient(network),
	}
}

type multiRaftConn struct {
	*client.Client
}

func (c *multiRaftConn) Connect(ctx context.Context, spec runtime.ConnSpec) error {
	var config multiraftv1.DriverConfig
	if err := spec.UnmarshalConfig(&config); err != nil {
		return err
	}
	return c.Client.Connect(ctx, config)
}

func (c *multiRaftConn) Configure(ctx context.Context, spec runtime.ConnSpec) error {
	var config multiraftv1.DriverConfig
	if err := spec.UnmarshalConfig(&config); err != nil {
		return err
	}
	return c.Client.Configure(ctx, config)
}

func (c *multiRaftConn) NewCounter(spec runtime.PrimitiveSpec) (counterv1.CounterServer, error) {
	return counterv1server.NewCounterServer(c.Protocol, spec)
}

func (c *multiRaftConn) NewCounterMap(spec runtime.PrimitiveSpec) (countermapv1.CounterMapServer, error) {
	return countermapv1server.NewCounterMapServer(c.Protocol, spec)
}

func (c *multiRaftConn) NewIndexedMap(spec runtime.PrimitiveSpec) (indexedmapv1.IndexedMapServer, error) {
	return indexedmapv1server.NewIndexedMapServer(c.Protocol, spec)
}

func (c *multiRaftConn) NewLeaderElection(spec runtime.PrimitiveSpec) (electionv1.LeaderElectionServer, error) {
	return electionv1server.NewLeaderElectionServer(c.Protocol, spec)
}

func (c *multiRaftConn) NewLock(spec runtime.PrimitiveSpec) (lockv1.LockServer, error) {
	return lockv1server.NewLockServer(c.Protocol, spec)
}

func (c *multiRaftConn) NewMap(spec runtime.PrimitiveSpec) (mapv1.MapServer, error) {
	return mapv1server.NewMapServer(c.Protocol, spec)
}

func (c *multiRaftConn) NewMultiMap(spec runtime.PrimitiveSpec) (multimapv1.MultiMapServer, error) {
	return multimapv1server.NewMultiMapServer(c.Protocol, spec)
}

func (c *multiRaftConn) NewSet(spec runtime.PrimitiveSpec) (setv1.SetServer, error) {
	return setv1server.NewSetServer(c.Protocol, spec)
}

func (c *multiRaftConn) NewValue(spec runtime.PrimitiveSpec) (valuev1.ValueServer, error) {
	return valuev1server.NewValueServer(c.Protocol, spec)
}
