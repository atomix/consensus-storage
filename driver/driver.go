// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"context"
	multiraftv1 "github.com/atomix/multi-raft/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft/driver/pkg/client"
	"github.com/atomix/runtime/pkg/runtime"
)

const (
	name    = "MultiRaft"
	version = "v1beta1"
)

var Driver = runtime.NewDriver[*multiraftv1.ClusterConfig](name, version, func(ctx context.Context, config *multiraftv1.ClusterConfig) (runtime.Client, error) {
	client := client.NewClient()
	if err := client.Connect(ctx, config); err != nil {
		return nil, err
	}
	return client, nil
})
