// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"context"
	multiraftcounterv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/counter/v1"
	multiraftcountermapv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/countermap/v1"
	multiraftelectionv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/election/v1"
	multiraftlockv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/lock/v1"
	multiraftmapv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/map/v1"
	multiraftsetv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/set/v1"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	multiraftvaluev1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/value/v1"
	"github.com/atomix/multi-raft-storage/driver/pkg/client"
	counterv1server "github.com/atomix/multi-raft-storage/driver/pkg/primitives/counter/v1"
	countermapv1server "github.com/atomix/multi-raft-storage/driver/pkg/primitives/countermap/v1"
	electionv1server "github.com/atomix/multi-raft-storage/driver/pkg/primitives/election/v1"
	lockv1server "github.com/atomix/multi-raft-storage/driver/pkg/primitives/lock/v1"
	mapv1server "github.com/atomix/multi-raft-storage/driver/pkg/primitives/map/v1"
	setv1server "github.com/atomix/multi-raft-storage/driver/pkg/primitives/set/v1"
	valuev1server "github.com/atomix/multi-raft-storage/driver/pkg/primitives/value/v1"
	counterv1 "github.com/atomix/runtime/api/atomix/runtime/counter/v1"
	countermapv1 "github.com/atomix/runtime/api/atomix/runtime/countermap/v1"
	electionv1 "github.com/atomix/runtime/api/atomix/runtime/election/v1"
	lockv1 "github.com/atomix/runtime/api/atomix/runtime/lock/v1"
	mapv1 "github.com/atomix/runtime/api/atomix/runtime/map/v1"
	setv1 "github.com/atomix/runtime/api/atomix/runtime/set/v1"
	valuev1 "github.com/atomix/runtime/api/atomix/runtime/value/v1"
	"github.com/atomix/runtime/sdk/pkg/runtime"
)

const (
	name    = "MultiRaft"
	version = "v1beta1"
)

func New(network runtime.Network) runtime.Driver {
	return runtime.NewDriver[multiraftv1.DriverConfig](name, version, func(ctx context.Context, config multiraftv1.DriverConfig) (runtime.Conn, error) {
		client := client.NewClient(network)
		if err := client.Connect(ctx, config); err != nil {
			return nil, err
		}
		return runtime.NewConn[multiraftv1.DriverConfig](client,
			runtime.WithCounterFactory[multiraftcounterv1.CounterConfig](func(config multiraftcounterv1.CounterConfig) (counterv1.CounterServer, error) {
				return counterv1server.NewCounterServer(client.Protocol, config), nil
			}),
			runtime.WithCounterMapFactory[multiraftcountermapv1.CounterMapConfig](func(config multiraftcountermapv1.CounterMapConfig) (countermapv1.CounterMapServer, error) {
				return countermapv1server.NewCounterMapServer(client.Protocol), nil
			}),
			runtime.WithLeaderElectionFactory[multiraftelectionv1.LeaderElectionConfig](func(config multiraftelectionv1.LeaderElectionConfig) (electionv1.LeaderElectionServer, error) {
				return electionv1server.NewLeaderElectionServer(client.Protocol), nil
			}),
			runtime.WithLockFactory[multiraftlockv1.LockConfig](func(config multiraftlockv1.LockConfig) (lockv1.LockServer, error) {
				return lockv1server.NewLockServer(client.Protocol), nil
			}),
			runtime.WithMapFactory[multiraftmapv1.MapConfig](func(config multiraftmapv1.MapConfig) (mapv1.MapServer, error) {
				return mapv1server.NewMapServer(client.Protocol, config), nil
			}),
			runtime.WithSetFactory[multiraftsetv1.SetConfig](func(config multiraftsetv1.SetConfig) (setv1.SetServer, error) {
				return setv1server.NewSetServer(client.Protocol, config), nil
			}),
			runtime.WithValueFactory[multiraftvaluev1.ValueConfig](func(config multiraftvaluev1.ValueConfig) (valuev1.ValueServer, error) {
				return valuev1server.NewValueServer(client.Protocol, config), nil
			})), nil
	})
}
