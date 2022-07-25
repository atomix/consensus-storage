// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package protocol

import (
	"context"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/runtime/sdk/pkg/errors"
	streams "github.com/atomix/runtime/sdk/pkg/stream"
	"github.com/gogo/protobuf/proto"
)

type Protocol[I any, O any] interface {
	Command(ctx context.Context, input I, headers *multiraftv1.CommandRequestHeaders) (O, *multiraftv1.CommandResponseHeaders, error)
	StreamCommand(ctx context.Context, input I, headers *multiraftv1.CommandRequestHeaders, stream streams.WriteStream[*StreamCommandResponse[O]]) error
	Query(ctx context.Context, input I, headers *multiraftv1.QueryRequestHeaders) (O, *multiraftv1.QueryResponseHeaders, error)
	StreamQuery(ctx context.Context, input I, headers *multiraftv1.QueryRequestHeaders, stream streams.WriteStream[*StreamQueryResponse[O]]) error
}

type StreamResponse[O any, H proto.Message] struct {
	Headers H
	Output  O
}

type StreamCommandResponse[O any] StreamResponse[O, *multiraftv1.CommandResponseHeaders]

type StreamQueryResponse[O any] StreamResponse[O, *multiraftv1.QueryResponseHeaders]

func NewProtocol[I, O proto.Message](node *Node, codec Codec[I, O]) Protocol[I, O] {
	return &genericProtocol[I, O]{
		node:  node,
		codec: codec,
	}
}

type genericProtocol[I, O proto.Message] struct {
	node  *Node
	codec Codec[I, O]
}

func (p *genericProtocol[I, O]) Command(ctx context.Context, input I, inputHeaders *multiraftv1.CommandRequestHeaders) (O, *multiraftv1.CommandResponseHeaders, error) {
	var output O
	inputBytes, err := p.codec.EncodeInput(input)
	if err != nil {
		return output, nil, errors.NewInternal(err.Error())
	}
	outputBytes, outputHeaders, err := p.node.Command(ctx, inputBytes, inputHeaders)
	if err != nil {
		return output, nil, err
	}
	output, err = p.codec.DecodeOutput(outputBytes)
	if err != nil {
		return output, nil, err
	}
	return output, outputHeaders, nil
}

func (p *genericProtocol[I, O]) StreamCommand(ctx context.Context, input I, headers *multiraftv1.CommandRequestHeaders, stream streams.WriteStream[*StreamCommandResponse[O]]) error {
	bytes, err := proto.Marshal(input)
	if err != nil {
		return errors.NewInternal(err.Error())
	}
	return p.node.StreamCommand(ctx, bytes, headers, streams.NewEncodingStream[*StreamCommandResponse[[]byte], *StreamCommandResponse[O]](stream, func(response *StreamCommandResponse[[]byte], err error) (*StreamCommandResponse[O], error) {
		if err != nil {
			return nil, err
		}
		output, err := p.codec.DecodeOutput(response.Output)
		if err != nil {
			return nil, err
		}
		return &StreamCommandResponse[O]{
			Headers: response.Headers,
			Output:  output,
		}, nil
	}))
}

func (p *genericProtocol[I, O]) Query(ctx context.Context, input I, inputHeaders *multiraftv1.QueryRequestHeaders) (O, *multiraftv1.QueryResponseHeaders, error) {
	var output O
	inputBytes, err := p.codec.EncodeInput(input)
	if err != nil {
		return output, nil, errors.NewInternal(err.Error())
	}
	outputBytes, outputHeaders, err := p.node.Query(ctx, inputBytes, inputHeaders)
	if err != nil {
		return output, nil, err
	}
	output, err = p.codec.DecodeOutput(outputBytes)
	if err != nil {
		return output, nil, err
	}
	return output, outputHeaders, nil
}

func (p *genericProtocol[I, O]) StreamQuery(ctx context.Context, input I, headers *multiraftv1.QueryRequestHeaders, stream streams.WriteStream[*StreamQueryResponse[O]]) error {
	bytes, err := proto.Marshal(input)
	if err != nil {
		return errors.NewInternal(err.Error())
	}
	return p.node.StreamQuery(ctx, bytes, headers, streams.NewEncodingStream[*StreamQueryResponse[[]byte], *StreamQueryResponse[O]](stream, func(response *StreamQueryResponse[[]byte], err error) (*StreamQueryResponse[O], error) {
		if err != nil {
			return nil, err
		}
		output, err := p.codec.DecodeOutput(response.Output)
		if err != nil {
			return nil, err
		}
		return &StreamQueryResponse[O]{
			Headers: response.Headers,
			Output:  output,
		}, nil
	}))
}
