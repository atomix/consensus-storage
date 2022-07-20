// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"fmt"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/atomix/runtime/sdk/pkg/grpc/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"sync"
)

func newPartitionClient(client *Client, id multiraftv1.PartitionID) *PartitionClient {
	return &PartitionClient{
		client: client,
		id:     id,
	}
}

type PartitionClient struct {
	client    *Client
	id        multiraftv1.PartitionID
	state     *PartitionState
	watchers  map[int]chan<- PartitionState
	watcherID int
	conn      *grpc.ClientConn
	session   *SessionClient
	mu        sync.RWMutex
}

func (p *PartitionClient) ID() multiraftv1.PartitionID {
	return p.id
}

func (p *PartitionClient) GetSession(ctx context.Context) (*SessionClient, error) {
	p.mu.RLock()
	session := p.session
	p.mu.RUnlock()
	if session != nil {
		return session, nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	if p.session != nil {
		return p.session, nil
	}
	if p.conn == nil {
		return nil, errors.NewUnavailable("not connected")
	}

	session = newSessionClient(p, p.conn)
	if err := session.open(ctx); err != nil {
		return nil, err
	}
	p.session = session
	return session, nil
}

func (p *PartitionClient) connect(ctx context.Context, config *multiraftv1.PartitionConfig) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	address := fmt.Sprintf("%s:///%s:%d", resolverName, config.Host, config.Port)
	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"loadBalancingPolicy":"%s"}`, resolverName)),
		grpc.WithResolvers(newResolver(p)),
		grpc.WithContextDialer(p.client.network.Connect),
		grpc.WithUnaryInterceptor(retry.RetryingUnaryClientInterceptor(retry.WithRetryOn(codes.Unavailable))),
		grpc.WithStreamInterceptor(retry.RetryingStreamClientInterceptor(retry.WithRetryOn(codes.Unavailable))))
	if err != nil {
		return errors.FromProto(err)
	}
	p.conn = conn
	return nil
}

type PartitionState struct {
	Leader    string
	Followers []string
}
