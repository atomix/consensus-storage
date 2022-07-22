// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package node

import (
	"context"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/node/manager"
	"github.com/atomix/multi-raft-storage/node/pkg/primitive"
	"github.com/atomix/runtime/sdk/pkg/errors"
	streams "github.com/atomix/runtime/sdk/pkg/stream"
)

func newProtocol(manager *manager.NodeManager) primitive.Protocol {
	return &primitiveProtocol{
		manager: manager,
	}
}

type primitiveProtocol struct {
	manager *manager.NodeManager
}

func (p *primitiveProtocol) Command(ctx context.Context, input []byte, headers *multiraftv1.CommandRequestHeaders) ([]byte, *multiraftv1.CommandResponseHeaders, error) {
	protocol := p.manager.Protocol()
	if protocol == nil {
		return nil, nil, errors.NewUnavailable("node not bootstrapped")
	}
	return protocol.Command(ctx, input, headers)
}

func (p *primitiveProtocol) StreamCommand(ctx context.Context, input []byte, headers *multiraftv1.CommandRequestHeaders, stream streams.WriteStream[*primitive.StreamResponse[*multiraftv1.CommandResponseHeaders]]) error {
	protocol := p.manager.Protocol()
	if protocol == nil {
		return errors.NewUnavailable("node not bootstrapped")
	}
	return protocol.StreamCommand(ctx, input, headers, stream)
}

func (p *primitiveProtocol) Query(ctx context.Context, input []byte, headers *multiraftv1.QueryRequestHeaders) ([]byte, *multiraftv1.QueryResponseHeaders, error) {
	protocol := p.manager.Protocol()
	if protocol == nil {
		return nil, nil, errors.NewUnavailable("node not bootstrapped")
	}
	return protocol.Query(ctx, input, headers)
}

func (p *primitiveProtocol) StreamQuery(ctx context.Context, input []byte, headers *multiraftv1.QueryRequestHeaders, stream streams.WriteStream[*primitive.StreamResponse[*multiraftv1.QueryResponseHeaders]]) error {
	protocol := p.manager.Protocol()
	if protocol == nil {
		return errors.NewUnavailable("node not bootstrapped")
	}
	return protocol.StreamQuery(ctx, input, headers, stream)
}
