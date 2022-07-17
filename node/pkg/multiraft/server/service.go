// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	multiraftv1 "github.com/atomix/multi-raft/api/atomix/multiraft/v1"
	"github.com/atomix/runtime/pkg/errors"
	"github.com/atomix/runtime/pkg/stream"
	"github.com/gogo/protobuf/proto"
	"github.com/lni/dragonboat/v3"
	"sync"
	"sync/atomic"
	"time"
)

const clientTimeout = 30 * time.Second

func newNodeService(id multiraftv1.NodeID, node *dragonboat.NodeHost) *NodeService {
	return &NodeService{
		id:         id,
		host:       node,
		partitions: make(map[multiraftv1.PartitionID]*PartitionService),
	}
}

type NodeService struct {
	id         multiraftv1.NodeID
	host       *dragonboat.NodeHost
	partitions map[multiraftv1.PartitionID]*PartitionService
	mu         sync.RWMutex
}

func (n *NodeService) ID() multiraftv1.NodeID {
	return n.id
}

func (n *NodeService) addPartition(partition *PartitionService) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.partitions[partition.id] = partition
}

func (n *NodeService) removePartition(partitionID multiraftv1.PartitionID) {
	n.mu.Lock()
	defer n.mu.Unlock()
	delete(n.partitions, partitionID)
}

func (n *NodeService) getPartition(partitionID multiraftv1.PartitionID) (*PartitionService, bool) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	partition, ok := n.partitions[partitionID]
	return partition, ok
}

func newPartitionService(node *NodeService, id multiraftv1.PartitionID) *PartitionService {
	return &PartitionService{
		node:    node,
		id:      id,
		streams: newStreamRegistry(),
	}
}

type PartitionService struct {
	node    *NodeService
	id      multiraftv1.PartitionID
	streams *streamRegistry
	leader  uint64
	term    uint64
}

func (p *PartitionService) ID() multiraftv1.PartitionID {
	return p.id
}

func (p *PartitionService) setLeader(term multiraftv1.Term, leader multiraftv1.NodeID) {
	atomic.StoreUint64(&p.term, uint64(term))
	atomic.StoreUint64(&p.leader, uint64(leader))
}

func (p *PartitionService) getLeader() (multiraftv1.Term, multiraftv1.NodeID) {
	return multiraftv1.Term(atomic.LoadUint64(&p.term)), multiraftv1.NodeID(atomic.LoadUint64(&p.leader))
}

func (n *NodeService) OpenSession(ctx context.Context, input *multiraftv1.OpenSessionInput, requestHeaders *multiraftv1.PartitionRequestHeaders) (*multiraftv1.OpenSessionOutput, *multiraftv1.PartitionResponseHeaders, error) {
	partition, ok := n.getPartition(requestHeaders.PartitionID)
	if !ok {
		return nil, nil, errors.NewForbidden("unknown partition %d", requestHeaders.PartitionID)
	}

	command := &multiraftv1.CommandInput{
		Timestamp: time.Now(),
		Input: &multiraftv1.CommandInput_OpenSession{
			OpenSession: input,
		},
	}
	output, err := partition.executeCommand(ctx, command)
	if err != nil {
		return nil, nil, err
	}
	responseHeaders := &multiraftv1.PartitionResponseHeaders{
		Index: output.Index,
	}
	return output.GetOpenSession(), responseHeaders, nil
}

func (n *NodeService) KeepAliveSession(ctx context.Context, input *multiraftv1.KeepAliveInput, requestHeaders *multiraftv1.PartitionRequestHeaders) (*multiraftv1.KeepAliveOutput, *multiraftv1.PartitionResponseHeaders, error) {
	partition, ok := n.getPartition(requestHeaders.PartitionID)
	if !ok {
		return nil, nil, errors.NewForbidden("unknown partition %d", requestHeaders.PartitionID)
	}

	command := &multiraftv1.CommandInput{
		Timestamp: time.Now(),
		Input: &multiraftv1.CommandInput_KeepAlive{
			KeepAlive: input,
		},
	}
	output, err := partition.executeCommand(ctx, command)
	if err != nil {
		return nil, nil, err
	}
	responseHeaders := &multiraftv1.PartitionResponseHeaders{
		Index: output.Index,
	}
	return output.GetKeepAlive(), responseHeaders, nil
}

func (n *NodeService) CloseSession(ctx context.Context, input *multiraftv1.CloseSessionInput, requestHeaders *multiraftv1.PartitionRequestHeaders) (*multiraftv1.CloseSessionOutput, *multiraftv1.PartitionResponseHeaders, error) {
	partition, ok := n.getPartition(requestHeaders.PartitionID)
	if !ok {
		return nil, nil, errors.NewForbidden("unknown partition %d", requestHeaders.PartitionID)
	}

	command := &multiraftv1.CommandInput{
		Timestamp: time.Now(),
		Input: &multiraftv1.CommandInput_CloseSession{
			CloseSession: input,
		},
	}
	output, err := partition.executeCommand(ctx, command)
	if err != nil {
		return nil, nil, err
	}
	responseHeaders := &multiraftv1.PartitionResponseHeaders{
		Index: output.Index,
	}
	return output.GetCloseSession(), responseHeaders, nil
}

func (n *NodeService) CreatePrimitive(ctx context.Context, primitiveSpec multiraftv1.PrimitiveSpec, requestHeaders *multiraftv1.SessionRequestHeaders) (*multiraftv1.CreatePrimitiveOutput, *multiraftv1.SessionResponseHeaders, error) {
	partition, ok := n.getPartition(requestHeaders.PartitionID)
	if !ok {
		return nil, nil, errors.NewForbidden("unknown partition %d", requestHeaders.PartitionID)
	}

	command := &multiraftv1.CommandInput{
		Timestamp: time.Now(),
		Input: &multiraftv1.CommandInput_SessionCommand{
			SessionCommand: &multiraftv1.SessionCommandInput{
				SessionID: requestHeaders.SessionID,
				Input: &multiraftv1.SessionCommandInput_CreatePrimitive{
					CreatePrimitive: &multiraftv1.CreatePrimitiveInput{
						PrimitiveSpec: primitiveSpec,
					},
				},
			},
		},
	}
	output, err := partition.executeCommand(ctx, command)
	if err != nil {
		return nil, nil, err
	}
	responseHeaders := &multiraftv1.SessionResponseHeaders{
		PartitionResponseHeaders: multiraftv1.PartitionResponseHeaders{
			Index: output.Index,
		},
	}
	return output.GetSessionCommand().GetCreatePrimitive(), responseHeaders, nil
}

func (n *NodeService) ClosePrimitive(ctx context.Context, primitiveID multiraftv1.PrimitiveID, requestHeaders *multiraftv1.SessionRequestHeaders) (*multiraftv1.ClosePrimitiveOutput, *multiraftv1.SessionResponseHeaders, error) {
	partition, ok := n.getPartition(requestHeaders.PartitionID)
	if !ok {
		return nil, nil, errors.NewForbidden("unknown partition %d", requestHeaders.PartitionID)
	}

	command := &multiraftv1.CommandInput{
		Timestamp: time.Now(),
		Input: &multiraftv1.CommandInput_SessionCommand{
			SessionCommand: &multiraftv1.SessionCommandInput{
				SessionID: requestHeaders.SessionID,
				Input: &multiraftv1.SessionCommandInput_ClosePrimitive{
					ClosePrimitive: &multiraftv1.ClosePrimitiveInput{
						PrimitiveID: primitiveID,
					},
				},
			},
		},
	}
	output, err := partition.executeCommand(ctx, command)
	if err != nil {
		return nil, nil, err
	}
	responseHeaders := &multiraftv1.SessionResponseHeaders{
		PartitionResponseHeaders: multiraftv1.PartitionResponseHeaders{
			Index: output.Index,
		},
	}
	return output.GetSessionCommand().GetClosePrimitive(), responseHeaders, nil
}

func (n *NodeService) ExecuteCommand(ctx context.Context, operationID multiraftv1.OperationID, inputBytes []byte, requestHeaders *multiraftv1.CommandRequestHeaders) ([]byte, *multiraftv1.CommandResponseHeaders, error) {
	partition, ok := n.getPartition(requestHeaders.PartitionID)
	if !ok {
		return nil, nil, errors.NewForbidden("unknown partition %d", requestHeaders.PartitionID)
	}

	command := &multiraftv1.CommandInput{
		Timestamp: time.Now(),
		Input: &multiraftv1.CommandInput_SessionCommand{
			SessionCommand: &multiraftv1.SessionCommandInput{
				SessionID: requestHeaders.SessionID,
				Input: &multiraftv1.SessionCommandInput_PrimitiveCommand{
					PrimitiveCommand: &multiraftv1.PrimitiveCommandInput{
						PrimitiveID: requestHeaders.PrimitiveID,
						SequenceNum: requestHeaders.SequenceNum,
						Operation: &multiraftv1.OperationInput{
							ID:      operationID,
							Payload: inputBytes,
						},
					},
				},
			},
		},
	}
	output, err := partition.executeCommand(ctx, command)
	if err != nil {
		return nil, nil, err
	}
	outputBytes := output.GetSessionCommand().GetPrimitiveCommand().GetOperation().Payload
	responseHeaders := &multiraftv1.CommandResponseHeaders{
		OperationResponseHeaders: multiraftv1.OperationResponseHeaders{
			PrimitiveResponseHeaders: multiraftv1.PrimitiveResponseHeaders{
				SessionResponseHeaders: multiraftv1.SessionResponseHeaders{
					PartitionResponseHeaders: multiraftv1.PartitionResponseHeaders{
						Index: output.Index,
					},
				},
			},
			Status:  getHeaderStatus(output.GetSessionCommand().GetPrimitiveCommand().GetOperation().Status),
			Message: output.GetSessionCommand().GetPrimitiveCommand().GetOperation().Message,
		},
		OutputSequenceNum: output.GetSessionCommand().GetPrimitiveCommand().SequenceNum,
	}
	return outputBytes, responseHeaders, nil
}

func (n *NodeService) ExecuteStreamCommand(ctx context.Context, operationID multiraftv1.OperationID, inputBytes []byte, requestHeaders *multiraftv1.CommandRequestHeaders, responseCh chan<- StreamCommandResponse) error {
	partition, ok := n.getPartition(requestHeaders.PartitionID)
	if !ok {
		return errors.NewForbidden("unknown partition %d", requestHeaders.PartitionID)
	}

	command := &multiraftv1.CommandInput{
		Timestamp: time.Now(),
		Input: &multiraftv1.CommandInput_SessionCommand{
			SessionCommand: &multiraftv1.SessionCommandInput{
				SessionID: requestHeaders.SessionID,
				Input: &multiraftv1.SessionCommandInput_PrimitiveCommand{
					PrimitiveCommand: &multiraftv1.PrimitiveCommandInput{
						PrimitiveID: requestHeaders.PrimitiveID,
						SequenceNum: requestHeaders.SequenceNum,
						Operation: &multiraftv1.OperationInput{
							ID:      operationID,
							Payload: inputBytes,
						},
					},
				},
			},
		},
	}

	outputCh := make(chan multiraftv1.CommandOutput)
	go func() {
		for output := range outputCh {
			responseCh <- StreamCommandResponse{
				Headers: multiraftv1.CommandResponseHeaders{
					OperationResponseHeaders: multiraftv1.OperationResponseHeaders{
						PrimitiveResponseHeaders: multiraftv1.PrimitiveResponseHeaders{
							SessionResponseHeaders: multiraftv1.SessionResponseHeaders{
								PartitionResponseHeaders: multiraftv1.PartitionResponseHeaders{
									Index: output.Index,
								},
							},
						},
						Status:  getHeaderStatus(output.GetSessionCommand().GetPrimitiveCommand().GetOperation().Status),
						Message: output.GetSessionCommand().GetPrimitiveCommand().GetOperation().Message,
					},
					OutputSequenceNum: output.GetSessionCommand().GetPrimitiveCommand().SequenceNum,
				},
				Payload: output.GetSessionCommand().GetPrimitiveCommand().GetOperation().Payload,
			}
		}
	}()
	return partition.executeStreamCommand(ctx, command, outputCh)
}

func (n *NodeService) ExecuteQuery(ctx context.Context, operationID multiraftv1.OperationID, inputBytes []byte, requestHeaders *multiraftv1.QueryRequestHeaders) ([]byte, *multiraftv1.QueryResponseHeaders, error) {
	partition, ok := n.getPartition(requestHeaders.PartitionID)
	if !ok {
		return nil, nil, errors.NewForbidden("unknown partition %d", requestHeaders.PartitionID)
	}

	query := &multiraftv1.QueryInput{
		MaxReceivedIndex: requestHeaders.MaxReceivedIndex,
		Input: &multiraftv1.QueryInput_SessionQuery{
			SessionQuery: &multiraftv1.SessionQueryInput{
				SessionID: requestHeaders.SessionID,
				Input: &multiraftv1.SessionQueryInput_PrimitiveQuery{
					PrimitiveQuery: &multiraftv1.PrimitiveQueryInput{
						PrimitiveID: requestHeaders.PrimitiveID,
						Operation: &multiraftv1.OperationInput{
							ID:      operationID,
							Payload: inputBytes,
						},
					},
				},
			},
		},
	}
	output, err := partition.executeQuery(ctx, query, requestHeaders.Sync)
	if err != nil {
		return nil, nil, err
	}
	outputBytes := output.GetSessionQuery().GetPrimitiveQuery().GetOperation().Payload
	responseHeaders := &multiraftv1.QueryResponseHeaders{
		OperationResponseHeaders: multiraftv1.OperationResponseHeaders{
			PrimitiveResponseHeaders: multiraftv1.PrimitiveResponseHeaders{
				SessionResponseHeaders: multiraftv1.SessionResponseHeaders{
					PartitionResponseHeaders: multiraftv1.PartitionResponseHeaders{
						Index: output.Index,
					},
				},
			},
			Status:  getHeaderStatus(output.GetSessionQuery().GetPrimitiveQuery().GetOperation().Status),
			Message: output.GetSessionQuery().GetPrimitiveQuery().GetOperation().Message,
		},
	}
	return outputBytes, responseHeaders, nil
}

func (n *NodeService) ExecuteStreamQuery(ctx context.Context, operationID multiraftv1.OperationID, inputBytes []byte, requestHeaders *multiraftv1.QueryRequestHeaders, responseCh chan<- StreamQueryResponse) error {
	partition, ok := n.getPartition(requestHeaders.PartitionID)
	if !ok {
		return errors.NewForbidden("unknown partition %d", requestHeaders.PartitionID)
	}

	query := &multiraftv1.QueryInput{
		MaxReceivedIndex: requestHeaders.MaxReceivedIndex,
		Input: &multiraftv1.QueryInput_SessionQuery{
			SessionQuery: &multiraftv1.SessionQueryInput{
				SessionID: requestHeaders.SessionID,
				Input: &multiraftv1.SessionQueryInput_PrimitiveQuery{
					PrimitiveQuery: &multiraftv1.PrimitiveQueryInput{
						PrimitiveID: requestHeaders.PrimitiveID,
						Operation: &multiraftv1.OperationInput{
							ID:      operationID,
							Payload: inputBytes,
						},
					},
				},
			},
		},
	}

	outputCh := make(chan multiraftv1.QueryOutput)
	go func() {
		for output := range outputCh {
			responseCh <- StreamQueryResponse{
				Headers: multiraftv1.QueryResponseHeaders{
					OperationResponseHeaders: multiraftv1.OperationResponseHeaders{
						PrimitiveResponseHeaders: multiraftv1.PrimitiveResponseHeaders{
							SessionResponseHeaders: multiraftv1.SessionResponseHeaders{
								PartitionResponseHeaders: multiraftv1.PartitionResponseHeaders{
									Index: output.Index,
								},
							},
						},
						Status:  getHeaderStatus(output.GetSessionQuery().GetPrimitiveQuery().GetOperation().Status),
						Message: output.GetSessionQuery().GetPrimitiveQuery().GetOperation().Message,
					},
				},
				Payload: output.GetSessionQuery().GetPrimitiveQuery().GetOperation().Payload,
			}
		}
	}()
	return partition.executeStreamQuery(ctx, query, outputCh, requestHeaders.Sync)
}

func (p *PartitionService) executeCommand(ctx context.Context, command *multiraftv1.CommandInput) (*multiraftv1.CommandOutput, error) {
	resultCh := make(chan stream.Result, 1)
	errCh := make(chan error, 1)
	go func() {
		if err := p.commitCommand(ctx, command, stream.NewChannelStream(resultCh)); err != nil {
			errCh <- err
		}
	}()

	select {
	case result, ok := <-resultCh:
		if !ok {
			err := errors.NewCanceled("stream closed")
			return nil, err
		}

		if result.Failed() {
			return nil, result.Error
		}

		return result.Value.(*multiraftv1.CommandOutput), nil
	case err := <-errCh:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (p *PartitionService) executeStreamCommand(ctx context.Context, command *multiraftv1.CommandInput, ch chan<- multiraftv1.CommandOutput) error {
	resultCh := make(chan stream.Result)
	errCh := make(chan error)

	stream := stream.NewBufferedStream()
	go func() {
		defer close(resultCh)
		for {
			result, ok := stream.Receive()
			if !ok {
				return
			}
			resultCh <- result
		}
	}()

	go func() {
		if err := p.commitCommand(ctx, command, stream); err != nil {
			errCh <- err
		}
	}()

	for {
		select {
		case result, ok := <-resultCh:
			if !ok {
				return nil
			}

			if result.Failed() {
				return result.Error
			}

			ch <- *result.Value.(*multiraftv1.CommandOutput)
		case err := <-errCh:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (p *PartitionService) commitCommand(ctx context.Context, input *multiraftv1.CommandInput, stream stream.WriteStream) error {
	term, leader := p.getLeader()
	if leader != p.node.id {
		return errors.NewUnavailable("not the leader")
	}

	streamID := p.streams.register(term, stream)
	defer p.streams.unregister(streamID)

	entry := &multiraftv1.RaftLogEntry{
		StreamID: streamID,
		Command:  *input,
	}

	bytes, err := proto.Marshal(entry)
	if err != nil {
		return errors.NewInternal("failed to marshal RaftLogEntry: %v", err)
	}

	ctx, cancel := context.WithTimeout(ctx, clientTimeout)
	defer cancel()
	if _, err := p.node.host.SyncPropose(ctx, p.node.host.GetNoOPSession(uint64(p.id)), bytes); err != nil {
		return wrapError(err)
	}
	return nil
}

func (p *PartitionService) executeQuery(ctx context.Context, query *multiraftv1.QueryInput, sync bool) (*multiraftv1.QueryOutput, error) {
	resultCh := make(chan stream.Result, 1)
	errCh := make(chan error, 1)
	go func() {
		if err := p.applyQuery(ctx, query, stream.NewChannelStream(resultCh), sync); err != nil {
			errCh <- err
		}
	}()

	select {
	case result, ok := <-resultCh:
		if !ok {
			err := errors.NewCanceled("stream closed")
			return nil, err
		}

		if result.Failed() {
			return nil, result.Error
		}

		return result.Value.(*multiraftv1.QueryOutput), nil
	case err := <-errCh:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (p *PartitionService) executeStreamQuery(ctx context.Context, query *multiraftv1.QueryInput, ch chan<- multiraftv1.QueryOutput, sync bool) error {
	resultCh := make(chan stream.Result)
	errCh := make(chan error)

	stream := stream.NewBufferedStream()
	go func() {
		defer close(resultCh)
		for {
			result, ok := stream.Receive()
			if !ok {
				return
			}
			resultCh <- result
		}
	}()

	go func() {
		if err := p.applyQuery(ctx, query, stream, sync); err != nil {
			errCh <- err
		}
	}()

	for {
		select {
		case result, ok := <-resultCh:
			if !ok {
				return nil
			}

			if result.Failed() {
				return result.Error
			}

			ch <- *result.Value.(*multiraftv1.QueryOutput)
		case err := <-errCh:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (p *PartitionService) applyQuery(ctx context.Context, input *multiraftv1.QueryInput, stream stream.WriteStream, sync bool) error {
	query := queryContext{
		input:  input,
		stream: stream,
	}
	if sync {
		ctx, cancel := context.WithTimeout(ctx, clientTimeout)
		defer cancel()
		if _, err := p.node.host.SyncRead(ctx, uint64(p.id), query); err != nil {
			return wrapError(err)
		}
	} else {
		if _, err := p.node.host.StaleRead(uint64(p.id), query); err != nil {
			return wrapError(err)
		}
	}
	return nil
}

func ExecuteCommand[I proto.Message, O proto.Message](node *NodeService) func(ctx context.Context, operationID multiraftv1.OperationID, input I, headers *multiraftv1.CommandRequestHeaders) (O, *multiraftv1.CommandResponseHeaders, error) {
	return func(ctx context.Context, operationID multiraftv1.OperationID, input I, requestHeaders *multiraftv1.CommandRequestHeaders) (O, *multiraftv1.CommandResponseHeaders, error) {
		var output O
		inputBytes, err := proto.Marshal(input)
		if err != nil {
			return output, nil, errors.NewInvalid(err.Error())
		}

		outputBytes, responseHeaders, err := node.ExecuteCommand(ctx, operationID, inputBytes, requestHeaders)
		if err != nil {
			return output, nil, err
		}

		err = proto.Unmarshal(outputBytes, output)
		if err != nil {
			return output, nil, err
		}
		return output, responseHeaders, nil
	}
}

func ExecuteStreamingCommand[I proto.Message, O proto.Message](ctx context.Context, input I, headers *multiraftv1.CommandRequestHeaders) {

}

func ExecuteQuery[I proto.Message, O proto.Message](node *NodeService) func(ctx context.Context, operationID multiraftv1.OperationID, input I, headers *multiraftv1.QueryRequestHeaders) (O, *multiraftv1.QueryResponseHeaders, error) {
	return func(ctx context.Context, operationID multiraftv1.OperationID, input I, requestHeaders *multiraftv1.QueryRequestHeaders) (O, *multiraftv1.QueryResponseHeaders, error) {
		var output O
		inputBytes, err := proto.Marshal(input)
		if err != nil {
			return output, nil, errors.NewInvalid(err.Error())
		}

		outputBytes, responseHeaders, err := node.ExecuteQuery(ctx, operationID, inputBytes, requestHeaders)
		if err != nil {
			return output, nil, err
		}

		err = proto.Unmarshal(outputBytes, output)
		if err != nil {
			return output, nil, err
		}
		return output, responseHeaders, nil
	}
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
