// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	multiraftv1 "github.com/atomix/multi-raft/api/atomix/multiraft/v1"
)

func newSessionClient(partition *PartitionClient) *SessionClient {
	return &SessionClient{
		partition: partition,
	}
}

type SessionClient struct {
	partition *PartitionClient
}

func (s *SessionClient) CreatePrimitive(ctx context.Context, spec multiraftv1.PrimitiveSpec) error {

}

func (s *SessionClient) GetPrimitive(ctx context.Context, name string) (*PrimitiveClient, error) {

}

func (s *SessionClient) ClosePrimitive(ctx context.Context, name string) error {

}
