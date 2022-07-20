// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package client

import (
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"hash/fnv"
	"sync"
)

func NewProtocol() *Protocol {
	return &Protocol{
		partitionIDs: make(map[multiraftv1.PartitionID]*PartitionClient),
	}
}

type Protocol struct {
	config       *multiraftv1.ClusterConfig
	partitions   []*PartitionClient
	partitionIDs map[multiraftv1.PartitionID]*PartitionClient
	mu           sync.RWMutex
}

func (p *Protocol) Partition(partitionID multiraftv1.PartitionID) *PartitionClient {
	return p.partitionIDs[partitionID]
}

func (p *Protocol) PartitionBy(partitionKey []byte) *PartitionClient {
	i, err := getPartitionIndex(partitionKey, len(p.partitions))
	if err != nil {
		panic(err)
	}
	return p.partitions[i]
}

func (p *Protocol) Partitions() []*PartitionClient {
	return p.partitions
}

// getPartitionIndex returns the index of the partition for the given key
func getPartitionIndex(key []byte, partitions int) (int, error) {
	h := fnv.New32a()
	if _, err := h.Write(key); err != nil {
		return 0, err
	}
	return int(h.Sum32() % uint32(partitions)), nil
}
