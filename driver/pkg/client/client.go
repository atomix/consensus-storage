// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	multiraftv1 "github.com/atomix/multi-raft/api/atomix/multiraft/v1"
	"github.com/atomix/runtime/pkg/logging"
	"sort"
	"sync"
)

var log = logging.GetLogger()

type Client struct {
	config       *multiraftv1.ClusterConfig
	partitions   []*Partition
	partitionIDs map[multiraftv1.PartitionID]*Partition
	mu           sync.RWMutex
}

func (c *Client) Connect(ctx context.Context, config *multiraftv1.ClusterConfig) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, partitionConfig := range config.Partitions {
		partition, ok := c.partitionIDs[partitionConfig.PartitionID]
		if !ok {
			partition = newPartition(c, partitionConfig.PartitionID)
			if err := partition.connect(ctx, &partitionConfig); err != nil {
				return err
			}
			c.partitionIDs[partition.id] = partition
			c.partitions = append(c.partitions, partition)
		}
	}

	sort.Slice(c.partitions, func(i, j int) bool {
		return c.partitions[i].id < c.partitions[j].id
	})

	c.config = config
	return nil
}

func (c *Client) Configure(ctx context.Context, config *multiraftv1.ClusterConfig) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config = config
	return nil
}

func (c *Client) Close(ctx context.Context) error {
	return nil
}
