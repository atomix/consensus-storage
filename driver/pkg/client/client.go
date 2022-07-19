// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	multiraftv1 "github.com/atomix/multi-raft/api/atomix/multiraft/v1"
	counterserver "github.com/atomix/multi-raft/driver/pkg/client/counter/v1"
	mapserver "github.com/atomix/multi-raft/driver/pkg/client/map/v1"
	counterv1 "github.com/atomix/runtime/api/atomix/runtime/counter/v1"
	mapv1 "github.com/atomix/runtime/api/atomix/runtime/map/v1"
	"github.com/atomix/runtime/pkg/logging"
	"hash/fnv"
	"sort"
	"sync"
)

var log = logging.GetLogger()

func NewClient() *Client {
	return &Client{
		partitionIDs: make(map[multiraftv1.PartitionID]*PartitionClient),
	}
}

type Client struct {
	config       *multiraftv1.ClusterConfig
	partitions   []*PartitionClient
	partitionIDs map[multiraftv1.PartitionID]*PartitionClient
	mu           sync.RWMutex
}

func (c *Client) GetCounter() counterv1.CounterServer {
	return counterserver.NewServer(c)
}

func (c *Client) GetMap() mapv1.MapServer {
	return mapserver.NewServer(c)
}

func (c *Client) Connect(ctx context.Context, config *multiraftv1.ClusterConfig) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, partitionConfig := range config.Partitions {
		partition, ok := c.partitionIDs[partitionConfig.PartitionID]
		if !ok {
			partition = newPartitionClient(c, partitionConfig.PartitionID)
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

func (c *Client) Partition(partitionID multiraftv1.PartitionID) *PartitionClient {
	return c.partitionIDs[partitionID]
}

func (c *Client) PartitionBy(partitionKey []byte) *PartitionClient {
	i, err := getPartitionIndex(partitionKey, len(c.partitions))
	if err != nil {
		panic(err)
	}
	return c.partitions[i]
}

func (c *Client) Partitions() []*PartitionClient {
	return c.partitions
}

func (c *Client) Close(ctx context.Context) error {
	return nil
}

// getPartitionIndex returns the index of the partition for the given key
func getPartitionIndex(key []byte, partitions int) (int, error) {
	h := fnv.New32a()
	if _, err := h.Write(key); err != nil {
		return 0, err
	}
	return int(h.Sum32() % uint32(partitions)), nil
}
