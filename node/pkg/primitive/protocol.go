// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package primitive

import (
	"context"
	multiraftv1 "github.com/atomix/multi-raft/api/atomix/multiraft/v1"
	streams "github.com/atomix/runtime/pkg/stream"
	"github.com/gogo/protobuf/proto"
)

type Protocol interface {
	Command(ctx context.Context, input []byte, headers *multiraftv1.CommandRequestHeaders) ([]byte, *multiraftv1.CommandResponseHeaders, error)
	StreamCommand(ctx context.Context, input []byte, headers *multiraftv1.CommandRequestHeaders, stream streams.WriteStream[*StreamResponse[*multiraftv1.CommandResponseHeaders]]) error
	Query(ctx context.Context, input []byte, headers *multiraftv1.QueryRequestHeaders) ([]byte, *multiraftv1.QueryResponseHeaders, error)
	StreamQuery(ctx context.Context, input []byte, headers *multiraftv1.QueryRequestHeaders, stream streams.WriteStream[*StreamResponse[*multiraftv1.QueryResponseHeaders]]) error
}

type StreamResponse[T proto.Message] struct {
	Headers T
	Output  []byte
}
