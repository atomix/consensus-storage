// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package node

import (
	"context"
	multiraftv1 "github.com/atomix/multi-raft/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft/node/pkg/primitive"
	"github.com/atomix/runtime/sdk/pkg/errors"
	streams "github.com/atomix/runtime/sdk/pkg/stream"
	"github.com/lni/dragonboat/v3"
	"sync"
	"time"
)

const clientTimeout = 30 * time.Second

func newProtocol(registry *primitive.Registry, options Options) *Protocol {
	protocol := &Protocol{
		id:         options.NodeID,
		partitions: make(map[multiraftv1.PartitionID]*Partition),
	}
	protocol.node = newManager(protocol, registry, options)
	return protocol
}

type Protocol struct {
	id         multiraftv1.NodeID
	node       *Manager
	partitions map[multiraftv1.PartitionID]*Partition
	mu         sync.RWMutex
}

func (p *Protocol) ID() multiraftv1.NodeID {
	return p.id
}

func (p *Protocol) addPartition(partition *Partition) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.partitions[partition.id] = partition
}

func (p *Protocol) removePartition(partitionID multiraftv1.PartitionID) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.partitions, partitionID)
}

func (p *Protocol) getPartition(partitionID multiraftv1.PartitionID) (*Partition, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	partition, ok := p.partitions[partitionID]
	return partition, ok
}

func (p *Protocol) OpenSession(ctx context.Context, input *multiraftv1.OpenSessionInput, requestHeaders *multiraftv1.PartitionRequestHeaders) (*multiraftv1.OpenSessionOutput, *multiraftv1.PartitionResponseHeaders, error) {
	partition, ok := p.getPartition(requestHeaders.PartitionID)
	if !ok {
		return nil, nil, errors.NewForbidden("unknown partition %d", requestHeaders.PartitionID)
	}

	command := &multiraftv1.CommandInput{
		Timestamp: time.Now(),
		Input: &multiraftv1.CommandInput_OpenSession{
			OpenSession: input,
		},
	}
	output, err := partition.Command(ctx, command)
	if err != nil {
		return nil, nil, err
	}
	responseHeaders := &multiraftv1.PartitionResponseHeaders{
		Index: output.Index,
	}
	return output.GetOpenSession(), responseHeaders, nil
}

func (p *Protocol) KeepAliveSession(ctx context.Context, input *multiraftv1.KeepAliveInput, requestHeaders *multiraftv1.PartitionRequestHeaders) (*multiraftv1.KeepAliveOutput, *multiraftv1.PartitionResponseHeaders, error) {
	partition, ok := p.getPartition(requestHeaders.PartitionID)
	if !ok {
		return nil, nil, errors.NewForbidden("unknown partition %d", requestHeaders.PartitionID)
	}

	command := &multiraftv1.CommandInput{
		Timestamp: time.Now(),
		Input: &multiraftv1.CommandInput_KeepAlive{
			KeepAlive: input,
		},
	}
	output, err := partition.Command(ctx, command)
	if err != nil {
		return nil, nil, err
	}
	responseHeaders := &multiraftv1.PartitionResponseHeaders{
		Index: output.Index,
	}
	return output.GetKeepAlive(), responseHeaders, nil
}

func (p *Protocol) CloseSession(ctx context.Context, input *multiraftv1.CloseSessionInput, requestHeaders *multiraftv1.PartitionRequestHeaders) (*multiraftv1.CloseSessionOutput, *multiraftv1.PartitionResponseHeaders, error) {
	partition, ok := p.getPartition(requestHeaders.PartitionID)
	if !ok {
		return nil, nil, errors.NewForbidden("unknown partition %d", requestHeaders.PartitionID)
	}

	command := &multiraftv1.CommandInput{
		Timestamp: time.Now(),
		Input: &multiraftv1.CommandInput_CloseSession{
			CloseSession: input,
		},
	}
	output, err := partition.Command(ctx, command)
	if err != nil {
		return nil, nil, err
	}
	responseHeaders := &multiraftv1.PartitionResponseHeaders{
		Index: output.Index,
	}
	return output.GetCloseSession(), responseHeaders, nil
}

func (p *Protocol) CreatePrimitive(ctx context.Context, input *multiraftv1.CreatePrimitiveInput, requestHeaders *multiraftv1.CommandRequestHeaders) (*multiraftv1.CreatePrimitiveOutput, *multiraftv1.CommandResponseHeaders, error) {
	partition, ok := p.getPartition(requestHeaders.PartitionID)
	if !ok {
		return nil, nil, errors.NewForbidden("unknown partition %d", requestHeaders.PartitionID)
	}

	command := &multiraftv1.CommandInput{
		Timestamp: time.Now(),
		Input: &multiraftv1.CommandInput_SessionCommand{
			SessionCommand: &multiraftv1.SessionCommandInput{
				SessionID:   requestHeaders.SessionID,
				SequenceNum: requestHeaders.SequenceNum,
				Input: &multiraftv1.SessionCommandInput_CreatePrimitive{
					CreatePrimitive: input,
				},
			},
		},
	}
	output, err := partition.Command(ctx, command)
	if err != nil {
		return nil, nil, err
	}
	responseHeaders := &multiraftv1.CommandResponseHeaders{
		OperationResponseHeaders: multiraftv1.OperationResponseHeaders{
			PrimitiveResponseHeaders: multiraftv1.PrimitiveResponseHeaders{
				SessionResponseHeaders: multiraftv1.SessionResponseHeaders{
					PartitionResponseHeaders: multiraftv1.PartitionResponseHeaders{
						Index: output.Index,
					},
				},
			},
		},
		OutputSequenceNum: output.GetSessionCommand().SequenceNum,
	}
	return output.GetSessionCommand().GetCreatePrimitive(), responseHeaders, nil
}

func (p *Protocol) ClosePrimitive(ctx context.Context, input *multiraftv1.ClosePrimitiveInput, requestHeaders *multiraftv1.CommandRequestHeaders) (*multiraftv1.ClosePrimitiveOutput, *multiraftv1.CommandResponseHeaders, error) {
	partition, ok := p.getPartition(requestHeaders.PartitionID)
	if !ok {
		return nil, nil, errors.NewForbidden("unknown partition %d", requestHeaders.PartitionID)
	}

	command := &multiraftv1.CommandInput{
		Timestamp: time.Now(),
		Input: &multiraftv1.CommandInput_SessionCommand{
			SessionCommand: &multiraftv1.SessionCommandInput{
				SessionID:   requestHeaders.SessionID,
				SequenceNum: requestHeaders.SequenceNum,
				Input: &multiraftv1.SessionCommandInput_ClosePrimitive{
					ClosePrimitive: input,
				},
			},
		},
	}
	output, err := partition.Command(ctx, command)
	if err != nil {
		return nil, nil, err
	}
	responseHeaders := &multiraftv1.CommandResponseHeaders{
		OperationResponseHeaders: multiraftv1.OperationResponseHeaders{
			PrimitiveResponseHeaders: multiraftv1.PrimitiveResponseHeaders{
				SessionResponseHeaders: multiraftv1.SessionResponseHeaders{
					PartitionResponseHeaders: multiraftv1.PartitionResponseHeaders{
						Index: output.Index,
					},
				},
			},
		},
		OutputSequenceNum: output.GetSessionCommand().SequenceNum,
	}
	return output.GetSessionCommand().GetClosePrimitive(), responseHeaders, nil
}

func (p *Protocol) Command(ctx context.Context, inputBytes []byte, requestHeaders *multiraftv1.CommandRequestHeaders) ([]byte, *multiraftv1.CommandResponseHeaders, error) {
	partition, ok := p.getPartition(requestHeaders.PartitionID)
	if !ok {
		return nil, nil, errors.NewForbidden("unknown partition %d", requestHeaders.PartitionID)
	}

	command := &multiraftv1.CommandInput{
		Timestamp: time.Now(),
		Input: &multiraftv1.CommandInput_SessionCommand{
			SessionCommand: &multiraftv1.SessionCommandInput{
				SessionID:   requestHeaders.SessionID,
				SequenceNum: requestHeaders.SequenceNum,
				Input: &multiraftv1.SessionCommandInput_Operation{
					Operation: &multiraftv1.PrimitiveOperationInput{
						PrimitiveID: requestHeaders.PrimitiveID,
						OperationInput: multiraftv1.OperationInput{
							ID:      requestHeaders.OperationID,
							Payload: inputBytes,
						},
					},
				},
			},
		},
	}
	output, err := partition.Command(ctx, command)
	if err != nil {
		return nil, nil, err
	}
	outputBytes := output.GetSessionCommand().GetOperation().Payload
	responseHeaders := &multiraftv1.CommandResponseHeaders{
		OperationResponseHeaders: multiraftv1.OperationResponseHeaders{
			PrimitiveResponseHeaders: multiraftv1.PrimitiveResponseHeaders{
				SessionResponseHeaders: multiraftv1.SessionResponseHeaders{
					PartitionResponseHeaders: multiraftv1.PartitionResponseHeaders{
						Index: output.Index,
					},
				},
			},
			Status:  getHeaderStatus(output.GetSessionCommand().GetOperation().Status),
			Message: output.GetSessionCommand().GetOperation().Message,
		},
		OutputSequenceNum: output.GetSessionCommand().SequenceNum,
	}
	return outputBytes, responseHeaders, nil
}

func (p *Protocol) StreamCommand(ctx context.Context, inputBytes []byte, requestHeaders *multiraftv1.CommandRequestHeaders, stream streams.WriteStream[*primitive.StreamResponse[*multiraftv1.CommandResponseHeaders]]) error {
	partition, ok := p.getPartition(requestHeaders.PartitionID)
	if !ok {
		return errors.NewForbidden("unknown partition %d", requestHeaders.PartitionID)
	}

	command := &multiraftv1.CommandInput{
		Timestamp: time.Now(),
		Input: &multiraftv1.CommandInput_SessionCommand{
			SessionCommand: &multiraftv1.SessionCommandInput{
				SessionID:   requestHeaders.SessionID,
				SequenceNum: requestHeaders.SequenceNum,
				Input: &multiraftv1.SessionCommandInput_Operation{
					Operation: &multiraftv1.PrimitiveOperationInput{
						PrimitiveID: requestHeaders.PrimitiveID,
						OperationInput: multiraftv1.OperationInput{
							ID:      requestHeaders.OperationID,
							Payload: inputBytes,
						},
					},
				},
			},
		},
	}
	return partition.StreamCommand(ctx, command, streams.NewEncodingStream[*multiraftv1.CommandOutput, *primitive.StreamResponse[*multiraftv1.CommandResponseHeaders]](stream, func(output *multiraftv1.CommandOutput, err error) (*primitive.StreamResponse[*multiraftv1.CommandResponseHeaders], error) {
		if err != nil {
			return nil, err
		}
		return &primitive.StreamResponse[*multiraftv1.CommandResponseHeaders]{
			Headers: &multiraftv1.CommandResponseHeaders{
				OperationResponseHeaders: multiraftv1.OperationResponseHeaders{
					PrimitiveResponseHeaders: multiraftv1.PrimitiveResponseHeaders{
						SessionResponseHeaders: multiraftv1.SessionResponseHeaders{
							PartitionResponseHeaders: multiraftv1.PartitionResponseHeaders{
								Index: output.Index,
							},
						},
					},
					Status:  getHeaderStatus(output.GetSessionCommand().GetOperation().Status),
					Message: output.GetSessionCommand().GetOperation().Message,
				},
				OutputSequenceNum: output.GetSessionCommand().SequenceNum,
			},
			Output: output.GetSessionCommand().GetOperation().Payload,
		}, nil
	}))
}

func (p *Protocol) Query(ctx context.Context, inputBytes []byte, requestHeaders *multiraftv1.QueryRequestHeaders) ([]byte, *multiraftv1.QueryResponseHeaders, error) {
	partition, ok := p.getPartition(requestHeaders.PartitionID)
	if !ok {
		return nil, nil, errors.NewForbidden("unknown partition %d", requestHeaders.PartitionID)
	}

	query := &multiraftv1.QueryInput{
		MaxReceivedIndex: requestHeaders.MaxReceivedIndex,
		Input: &multiraftv1.QueryInput_SessionQuery{
			SessionQuery: &multiraftv1.SessionQueryInput{
				SessionID: requestHeaders.SessionID,
				Input: &multiraftv1.SessionQueryInput_Operation{
					Operation: &multiraftv1.PrimitiveOperationInput{
						PrimitiveID: requestHeaders.PrimitiveID,
						OperationInput: multiraftv1.OperationInput{
							ID:      requestHeaders.OperationID,
							Payload: inputBytes,
						},
					},
				},
			},
		},
	}
	output, err := partition.Query(ctx, query)
	if err != nil {
		return nil, nil, err
	}
	outputBytes := output.GetSessionQuery().GetOperation().Payload
	responseHeaders := &multiraftv1.QueryResponseHeaders{
		OperationResponseHeaders: multiraftv1.OperationResponseHeaders{
			PrimitiveResponseHeaders: multiraftv1.PrimitiveResponseHeaders{
				SessionResponseHeaders: multiraftv1.SessionResponseHeaders{
					PartitionResponseHeaders: multiraftv1.PartitionResponseHeaders{
						Index: output.Index,
					},
				},
			},
			Status:  getHeaderStatus(output.GetSessionQuery().GetOperation().Status),
			Message: output.GetSessionQuery().GetOperation().Message,
		},
	}
	return outputBytes, responseHeaders, nil
}

func (p *Protocol) StreamQuery(ctx context.Context, inputBytes []byte, requestHeaders *multiraftv1.QueryRequestHeaders, stream streams.WriteStream[*primitive.StreamResponse[*multiraftv1.QueryResponseHeaders]]) error {
	partition, ok := p.getPartition(requestHeaders.PartitionID)
	if !ok {
		return errors.NewForbidden("unknown partition %d", requestHeaders.PartitionID)
	}

	query := &multiraftv1.QueryInput{
		MaxReceivedIndex: requestHeaders.MaxReceivedIndex,
		Input: &multiraftv1.QueryInput_SessionQuery{
			SessionQuery: &multiraftv1.SessionQueryInput{
				SessionID: requestHeaders.SessionID,
				Input: &multiraftv1.SessionQueryInput_Operation{
					Operation: &multiraftv1.PrimitiveOperationInput{
						PrimitiveID: requestHeaders.PrimitiveID,
						OperationInput: multiraftv1.OperationInput{
							ID:      requestHeaders.OperationID,
							Payload: inputBytes,
						},
					},
				},
			},
		},
	}
	return partition.StreamQuery(ctx, query, streams.NewEncodingStream[*multiraftv1.QueryOutput, *primitive.StreamResponse[*multiraftv1.QueryResponseHeaders]](stream, func(output *multiraftv1.QueryOutput, err error) (*primitive.StreamResponse[*multiraftv1.QueryResponseHeaders], error) {
		if err != nil {
			return nil, err
		}
		return &primitive.StreamResponse[*multiraftv1.QueryResponseHeaders]{
			Headers: &multiraftv1.QueryResponseHeaders{
				OperationResponseHeaders: multiraftv1.OperationResponseHeaders{
					PrimitiveResponseHeaders: multiraftv1.PrimitiveResponseHeaders{
						SessionResponseHeaders: multiraftv1.SessionResponseHeaders{
							PartitionResponseHeaders: multiraftv1.PartitionResponseHeaders{
								Index: output.Index,
							},
						},
					},
					Status:  getHeaderStatus(output.GetSessionQuery().GetOperation().Status),
					Message: output.GetSessionQuery().GetOperation().Message,
				},
			},
			Output: output.GetSessionQuery().GetOperation().Payload,
		}, nil
	}))
}

type StreamCommandResponse struct {
	Headers multiraftv1.CommandResponseHeaders
	Payload []byte
}

type StreamQueryResponse struct {
	Headers multiraftv1.QueryResponseHeaders
	Payload []byte
}

func getHeaderStatus(status multiraftv1.OperationOutput_Status) multiraftv1.OperationResponseHeaders_Status {
	switch status {
	case multiraftv1.OperationOutput_UNKNOWN:
		return multiraftv1.OperationResponseHeaders_UNKNOWN
	case multiraftv1.OperationOutput_OK:
		return multiraftv1.OperationResponseHeaders_OK
	case multiraftv1.OperationOutput_ERROR:
		return multiraftv1.OperationResponseHeaders_ERROR
	case multiraftv1.OperationOutput_CANCELED:
		return multiraftv1.OperationResponseHeaders_CANCELED
	case multiraftv1.OperationOutput_NOT_FOUND:
		return multiraftv1.OperationResponseHeaders_NOT_FOUND
	case multiraftv1.OperationOutput_ALREADY_EXISTS:
		return multiraftv1.OperationResponseHeaders_ALREADY_EXISTS
	case multiraftv1.OperationOutput_UNAUTHORIZED:
		return multiraftv1.OperationResponseHeaders_UNAUTHORIZED
	case multiraftv1.OperationOutput_FORBIDDEN:
		return multiraftv1.OperationResponseHeaders_FORBIDDEN
	case multiraftv1.OperationOutput_CONFLICT:
		return multiraftv1.OperationResponseHeaders_CONFLICT
	case multiraftv1.OperationOutput_INVALID:
		return multiraftv1.OperationResponseHeaders_INVALID
	case multiraftv1.OperationOutput_UNAVAILABLE:
		return multiraftv1.OperationResponseHeaders_UNAVAILABLE
	case multiraftv1.OperationOutput_NOT_SUPPORTED:
		return multiraftv1.OperationResponseHeaders_NOT_SUPPORTED
	case multiraftv1.OperationOutput_TIMEOUT:
		return multiraftv1.OperationResponseHeaders_TIMEOUT
	case multiraftv1.OperationOutput_INTERNAL:
		return multiraftv1.OperationResponseHeaders_INTERNAL
	default:
		return multiraftv1.OperationResponseHeaders_UNKNOWN
	}
}

func wrapError(err error) error {
	switch err {
	case dragonboat.ErrClusterNotFound,
		dragonboat.ErrClusterNotBootstrapped,
		dragonboat.ErrClusterNotInitialized,
		dragonboat.ErrClusterNotReady,
		dragonboat.ErrClusterClosed:
		return errors.NewUnavailable(err.Error())
	case dragonboat.ErrSystemBusy,
		dragonboat.ErrBadKey:
		return errors.NewUnavailable(err.Error())
	case dragonboat.ErrClosed,
		dragonboat.ErrNodeRemoved:
		return errors.NewUnavailable(err.Error())
	case dragonboat.ErrInvalidSession,
		dragonboat.ErrInvalidTarget,
		dragonboat.ErrInvalidAddress,
		dragonboat.ErrInvalidOperation:
		return errors.NewInvalid(err.Error())
	case dragonboat.ErrPayloadTooBig,
		dragonboat.ErrTimeoutTooSmall:
		return errors.NewForbidden(err.Error())
	case dragonboat.ErrDeadlineNotSet,
		dragonboat.ErrInvalidDeadline:
		return errors.NewInternal(err.Error())
	case dragonboat.ErrDirNotExist:
		return errors.NewInternal(err.Error())
	case dragonboat.ErrTimeout:
		return errors.NewTimeout(err.Error())
	case dragonboat.ErrCanceled:
		return errors.NewCanceled(err.Error())
	case dragonboat.ErrRejected:
		return errors.NewForbidden(err.Error())
	default:
		return errors.NewUnknown(err.Error())
	}
}
