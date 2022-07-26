// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package protocol

import (
	"context"
	"fmt"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine"
	"github.com/atomix/runtime/sdk/pkg/async"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/atomix/runtime/sdk/pkg/logging"
	streams "github.com/atomix/runtime/sdk/pkg/stream"
	"github.com/lni/dragonboat/v3"
	raftconfig "github.com/lni/dragonboat/v3/config"
	"sync"
	"time"
)

var log = logging.GetLogger()

const (
	defaultDataDir = "/var/lib/atomix/data"
)

const clientTimeout = 30 * time.Second

func NewNode(registry *statemachine.PrimitiveTypeRegistry, config multiraftv1.NodeConfig) *Node {
	var rtt uint64 = 250
	if config.HeartbeatPeriod != nil {
		rtt = uint64(config.HeartbeatPeriod.Milliseconds())
	}

	dataDir := config.DataDir
	if dataDir == "" {
		dataDir = defaultDataDir
	}

	node := &Node{
		id:         config.NodeID,
		registry:   registry,
		config:     config,
		partitions: make(map[multiraftv1.PartitionID]*Partition),
		listeners:  make(map[int]chan<- multiraftv1.NodeEvent),
	}

	listener := newEventListener(node)
	address := fmt.Sprintf("%s:%d", config.Host, config.Port)
	nodeConfig := raftconfig.NodeHostConfig{
		WALDir:              dataDir,
		NodeHostDir:         dataDir,
		RTTMillisecond:      rtt,
		RaftAddress:         address,
		RaftEventListener:   listener,
		SystemEventListener: listener,
	}

	host, err := dragonboat.NewNodeHost(nodeConfig)
	if err != nil {
		panic(err)
	}

	node.host = host
	return node
}

type Node struct {
	id         multiraftv1.NodeID
	host       *dragonboat.NodeHost
	config     multiraftv1.NodeConfig
	cluster    *multiraftv1.ClusterConfig
	registry   *statemachine.PrimitiveTypeRegistry
	partitions map[multiraftv1.PartitionID]*Partition
	listeners  map[int]chan<- multiraftv1.NodeEvent
	listenerID int
	mu         sync.RWMutex
}

func (n *Node) ID() multiraftv1.NodeID {
	return n.id
}

func (n *Node) publish(event *multiraftv1.NodeEvent) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	log.Infow("Publish NodeEvent",
		logging.Stringer("NodeEvent", event))
	for _, listener := range n.listeners {
		listener <- *event
	}
}

func (n *Node) Partition(partitionID multiraftv1.PartitionID) (*Partition, bool) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	partition, ok := n.partitions[partitionID]
	return partition, ok
}

func (n *Node) Watch(ctx context.Context, ch chan<- multiraftv1.NodeEvent) {
	n.listenerID++
	id := n.listenerID
	n.mu.Lock()
	n.listeners[id] = ch
	n.mu.Unlock()

	go func() {
		<-ctx.Done()
		n.mu.Lock()
		close(ch)
		delete(n.listeners, id)
		n.mu.Unlock()
	}()
}

func (n *Node) Bootstrap(config multiraftv1.ClusterConfig) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.cluster = &config

	if len(n.partitions) > 0 {
		return errors.NewConflict("node is already bootstrapped")
	}

	for _, partitionConfig := range config.Partitions {
		n.partitions[partitionConfig.PartitionID] = newPartition(partitionConfig.PartitionID, n)
	}

	return async.IterAsync(len(config.Partitions), func(i int) error {
		partitionConfig := config.Partitions[i]
		return n.partitions[partitionConfig.PartitionID].bootstrap(config, partitionConfig)
	})
}

func (n *Node) Join(config multiraftv1.PartitionConfig) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.cluster == nil {
		return errors.NewForbidden("node has not yet been bootstrapped")
	}

	partition, ok := n.partitions[config.PartitionID]
	if !ok {
		return errors.NewFault("partition %d not found", config.PartitionID)
	}
	return partition.join(*n.cluster, config)
}

func (n *Node) Leave(partitionID multiraftv1.PartitionID) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.cluster == nil {
		return errors.NewForbidden("node has not yet been bootstrapped")
	}

	partition, ok := n.partitions[partitionID]
	if !ok {
		return errors.NewFault("partition %d not found", partitionID)
	}
	return partition.leave()
}

func (n *Node) Shutdown() error {
	n.host.Stop()
	return nil
}

func (n *Node) OpenSession(ctx context.Context, input *multiraftv1.OpenSessionInput, requestHeaders *multiraftv1.PartitionRequestHeaders) (*multiraftv1.OpenSessionOutput, *multiraftv1.PartitionResponseHeaders, error) {
	partition, ok := n.Partition(requestHeaders.PartitionID)
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

func (n *Node) KeepAliveSession(ctx context.Context, input *multiraftv1.KeepAliveInput, requestHeaders *multiraftv1.PartitionRequestHeaders) (*multiraftv1.KeepAliveOutput, *multiraftv1.PartitionResponseHeaders, error) {
	partition, ok := n.Partition(requestHeaders.PartitionID)
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

func (n *Node) CloseSession(ctx context.Context, input *multiraftv1.CloseSessionInput, requestHeaders *multiraftv1.PartitionRequestHeaders) (*multiraftv1.CloseSessionOutput, *multiraftv1.PartitionResponseHeaders, error) {
	partition, ok := n.Partition(requestHeaders.PartitionID)
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

func (n *Node) CreatePrimitive(ctx context.Context, input *multiraftv1.CreatePrimitiveInput, requestHeaders *multiraftv1.CommandRequestHeaders) (*multiraftv1.CreatePrimitiveOutput, *multiraftv1.CommandResponseHeaders, error) {
	partition, ok := n.Partition(requestHeaders.PartitionID)
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

func (n *Node) ClosePrimitive(ctx context.Context, input *multiraftv1.ClosePrimitiveInput, requestHeaders *multiraftv1.CommandRequestHeaders) (*multiraftv1.ClosePrimitiveOutput, *multiraftv1.CommandResponseHeaders, error) {
	partition, ok := n.Partition(requestHeaders.PartitionID)
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

func (n *Node) Command(ctx context.Context, inputBytes []byte, requestHeaders *multiraftv1.CommandRequestHeaders) ([]byte, *multiraftv1.CommandResponseHeaders, error) {
	partition, ok := n.Partition(requestHeaders.PartitionID)
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
	responseHeaders := &multiraftv1.CommandResponseHeaders{
		OperationResponseHeaders: multiraftv1.OperationResponseHeaders{
			PrimitiveResponseHeaders: multiraftv1.PrimitiveResponseHeaders{
				SessionResponseHeaders: multiraftv1.SessionResponseHeaders{
					PartitionResponseHeaders: multiraftv1.PartitionResponseHeaders{
						Index: output.Index,
					},
				},
			},
			Status:  getHeaderStatus(output.GetSessionCommand().Failure),
			Message: getHeaderMessage(output.GetSessionCommand().Failure),
		},
		OutputSequenceNum: output.GetSessionCommand().SequenceNum,
	}
	if responseHeaders.Status != multiraftv1.OperationResponseHeaders_OK {
		return nil, responseHeaders, nil
	}
	return output.GetSessionCommand().GetOperation().Payload, responseHeaders, nil
}

func (n *Node) StreamCommand(ctx context.Context, inputBytes []byte, requestHeaders *multiraftv1.CommandRequestHeaders, stream streams.WriteStream[*StreamCommandResponse[[]byte]]) error {
	partition, ok := n.Partition(requestHeaders.PartitionID)
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
							Payload: inputBytes,
						},
					},
				},
			},
		},
	}
	return partition.StreamCommand(ctx, command, streams.NewEncodingStream[*multiraftv1.CommandOutput, *StreamCommandResponse[[]byte]](stream, func(output *multiraftv1.CommandOutput, err error) (*StreamCommandResponse[[]byte], error) {
		if err != nil {
			return nil, err
		}
		headers := &multiraftv1.CommandResponseHeaders{
			OperationResponseHeaders: multiraftv1.OperationResponseHeaders{
				PrimitiveResponseHeaders: multiraftv1.PrimitiveResponseHeaders{
					SessionResponseHeaders: multiraftv1.SessionResponseHeaders{
						PartitionResponseHeaders: multiraftv1.PartitionResponseHeaders{
							Index: output.Index,
						},
					},
				},
				Status:  getHeaderStatus(output.GetSessionCommand().Failure),
				Message: getHeaderMessage(output.GetSessionCommand().Failure),
			},
			OutputSequenceNum: output.GetSessionCommand().SequenceNum,
		}
		var payload []byte
		if headers.Status == multiraftv1.OperationResponseHeaders_OK {
			payload = output.GetSessionCommand().GetOperation().Payload
		}
		return &StreamCommandResponse[[]byte]{
			Headers: headers,
			Output:  payload,
		}, nil
	}))
}

func (n *Node) Query(ctx context.Context, inputBytes []byte, requestHeaders *multiraftv1.QueryRequestHeaders) ([]byte, *multiraftv1.QueryResponseHeaders, error) {
	partition, ok := n.Partition(requestHeaders.PartitionID)
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
	responseHeaders := &multiraftv1.QueryResponseHeaders{
		OperationResponseHeaders: multiraftv1.OperationResponseHeaders{
			PrimitiveResponseHeaders: multiraftv1.PrimitiveResponseHeaders{
				SessionResponseHeaders: multiraftv1.SessionResponseHeaders{
					PartitionResponseHeaders: multiraftv1.PartitionResponseHeaders{
						Index: output.Index,
					},
				},
			},
			Status:  getHeaderStatus(output.GetSessionQuery().Failure),
			Message: getHeaderMessage(output.GetSessionQuery().Failure),
		},
	}
	if responseHeaders.Status != multiraftv1.OperationResponseHeaders_OK {
		return nil, responseHeaders, nil
	}
	return output.GetSessionQuery().GetOperation().Payload, responseHeaders, nil
}

func (n *Node) StreamQuery(ctx context.Context, inputBytes []byte, requestHeaders *multiraftv1.QueryRequestHeaders, stream streams.WriteStream[*StreamQueryResponse[[]byte]]) error {
	partition, ok := n.Partition(requestHeaders.PartitionID)
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
							Payload: inputBytes,
						},
					},
				},
			},
		},
	}
	return partition.StreamQuery(ctx, query, streams.NewEncodingStream[*multiraftv1.QueryOutput, *StreamQueryResponse[[]byte]](stream, func(output *multiraftv1.QueryOutput, err error) (*StreamQueryResponse[[]byte], error) {
		if err != nil {
			return nil, err
		}
		headers := &multiraftv1.QueryResponseHeaders{
			OperationResponseHeaders: multiraftv1.OperationResponseHeaders{
				PrimitiveResponseHeaders: multiraftv1.PrimitiveResponseHeaders{
					SessionResponseHeaders: multiraftv1.SessionResponseHeaders{
						PartitionResponseHeaders: multiraftv1.PartitionResponseHeaders{
							Index: output.Index,
						},
					},
				},
				Status:  getHeaderStatus(output.GetSessionQuery().Failure),
				Message: getHeaderMessage(output.GetSessionQuery().Failure),
			},
		}
		var payload []byte
		if headers.Status == multiraftv1.OperationResponseHeaders_OK {
			payload = output.GetSessionQuery().GetOperation().Payload
		}
		return &StreamQueryResponse[[]byte]{
			Headers: headers,
			Output:  payload,
		}, nil
	}))
}

func getHeaderStatus(failure *multiraftv1.Failure) multiraftv1.OperationResponseHeaders_Status {
	if failure == nil {
		return multiraftv1.OperationResponseHeaders_OK
	}
	switch failure.Status {
	case multiraftv1.Failure_UNKNOWN:
		return multiraftv1.OperationResponseHeaders_UNKNOWN
	case multiraftv1.Failure_ERROR:
		return multiraftv1.OperationResponseHeaders_ERROR
	case multiraftv1.Failure_CANCELED:
		return multiraftv1.OperationResponseHeaders_CANCELED
	case multiraftv1.Failure_NOT_FOUND:
		return multiraftv1.OperationResponseHeaders_NOT_FOUND
	case multiraftv1.Failure_ALREADY_EXISTS:
		return multiraftv1.OperationResponseHeaders_ALREADY_EXISTS
	case multiraftv1.Failure_UNAUTHORIZED:
		return multiraftv1.OperationResponseHeaders_UNAUTHORIZED
	case multiraftv1.Failure_FORBIDDEN:
		return multiraftv1.OperationResponseHeaders_FORBIDDEN
	case multiraftv1.Failure_CONFLICT:
		return multiraftv1.OperationResponseHeaders_CONFLICT
	case multiraftv1.Failure_INVALID:
		return multiraftv1.OperationResponseHeaders_INVALID
	case multiraftv1.Failure_UNAVAILABLE:
		return multiraftv1.OperationResponseHeaders_UNAVAILABLE
	case multiraftv1.Failure_NOT_SUPPORTED:
		return multiraftv1.OperationResponseHeaders_NOT_SUPPORTED
	case multiraftv1.Failure_TIMEOUT:
		return multiraftv1.OperationResponseHeaders_TIMEOUT
	case multiraftv1.Failure_INTERNAL:
		return multiraftv1.OperationResponseHeaders_INTERNAL
	case multiraftv1.Failure_FAULT:
		return multiraftv1.OperationResponseHeaders_FAULT
	default:
		return multiraftv1.OperationResponseHeaders_UNKNOWN
	}
}

func getHeaderMessage(failure *multiraftv1.Failure) string {
	if failure != nil {
		return failure.Message
	}
	return ""
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
