// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"context"
	countermapv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/countermap/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/protocol"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/atomix/runtime/sdk/pkg/logging"
	streams "github.com/atomix/runtime/sdk/pkg/stream"
	"github.com/atomix/runtime/sdk/pkg/stringer"
	"github.com/gogo/protobuf/proto"
)

var log = logging.GetLogger()

const truncLen = 250

var counterMapCodec = protocol.NewCodec[*countermapv1.CounterMapInput, *countermapv1.CounterMapOutput](
	func(input *countermapv1.CounterMapInput) ([]byte, error) {
		return proto.Marshal(input)
	},
	func(bytes []byte) (*countermapv1.CounterMapOutput, error) {
		output := &countermapv1.CounterMapOutput{}
		if err := proto.Unmarshal(bytes, output); err != nil {
			return nil, err
		}
		return output, nil
	})

func NewCounterMapServer(node *protocol.Node) countermapv1.CounterMapServer {
	return &CounterMapServer{
		protocol: protocol.NewProtocol[*countermapv1.CounterMapInput, *countermapv1.CounterMapOutput](node, counterMapCodec),
	}
}

type CounterMapServer struct {
	protocol protocol.Protocol[*countermapv1.CounterMapInput, *countermapv1.CounterMapOutput]
}

func (s *CounterMapServer) Size(ctx context.Context, request *countermapv1.SizeRequest) (*countermapv1.SizeResponse, error) {
	log.Debugw("Size",
		logging.Stringer("SizeRequest", stringer.Truncate(request, truncLen)))
	input := &countermapv1.CounterMapInput{
		Input: &countermapv1.CounterMapInput_Size_{
			Size_: request.SizeInput,
		},
	}
	output, headers, err := s.protocol.Query(ctx, input, request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Size",
			logging.Stringer("SizeRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, err
	}
	response := &countermapv1.SizeResponse{
		Headers:    headers,
		SizeOutput: output.GetSize_(),
	}
	log.Debugw("Size",
		logging.Stringer("SizeRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("SizeResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *CounterMapServer) Set(ctx context.Context, request *countermapv1.SetRequest) (*countermapv1.SetResponse, error) {
	log.Debugw("Set",
		logging.Stringer("SetRequest", stringer.Truncate(request, truncLen)))
	input := &countermapv1.CounterMapInput{
		Input: &countermapv1.CounterMapInput_Set{
			Set: request.SetInput,
		},
	}
	output, headers, err := s.protocol.Command(ctx, input, request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Set",
			logging.Stringer("SetRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, err
	}
	response := &countermapv1.SetResponse{
		Headers:   headers,
		SetOutput: output.GetSet(),
	}
	log.Debugw("Set",
		logging.Stringer("SetRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("SetResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *CounterMapServer) Insert(ctx context.Context, request *countermapv1.InsertRequest) (*countermapv1.InsertResponse, error) {
	log.Debugw("Insert",
		logging.Stringer("InsertRequest", stringer.Truncate(request, truncLen)))
	input := &countermapv1.CounterMapInput{
		Input: &countermapv1.CounterMapInput_Insert{
			Insert: request.InsertInput,
		},
	}
	output, headers, err := s.protocol.Command(ctx, input, request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Insert",
			logging.Stringer("InsertRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, err
	}
	response := &countermapv1.InsertResponse{
		Headers:      headers,
		InsertOutput: output.GetInsert(),
	}
	log.Debugw("Insert",
		logging.Stringer("InsertRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("InsertResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *CounterMapServer) Update(ctx context.Context, request *countermapv1.UpdateRequest) (*countermapv1.UpdateResponse, error) {
	log.Debugw("Update",
		logging.Stringer("UpdateRequest", stringer.Truncate(request, truncLen)))
	input := &countermapv1.CounterMapInput{
		Input: &countermapv1.CounterMapInput_Update{
			Update: request.UpdateInput,
		},
	}
	output, headers, err := s.protocol.Command(ctx, input, request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Update",
			logging.Stringer("UpdateRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, err
	}
	response := &countermapv1.UpdateResponse{
		Headers:      headers,
		UpdateOutput: output.GetUpdate(),
	}
	log.Debugw("Update",
		logging.Stringer("UpdateRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("UpdateResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *CounterMapServer) Increment(ctx context.Context, request *countermapv1.IncrementRequest) (*countermapv1.IncrementResponse, error) {
	log.Debugw("Increment",
		logging.Stringer("IncrementRequest", stringer.Truncate(request, truncLen)))
	input := &countermapv1.CounterMapInput{
		Input: &countermapv1.CounterMapInput_Increment{
			Increment: request.IncrementInput,
		},
	}
	output, headers, err := s.protocol.Command(ctx, input, request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Increment",
			logging.Stringer("IncrementRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, err
	}
	response := &countermapv1.IncrementResponse{
		Headers:         headers,
		IncrementOutput: output.GetIncrement(),
	}
	log.Debugw("Increment",
		logging.Stringer("IncrementRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("IncrementResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *CounterMapServer) Decrement(ctx context.Context, request *countermapv1.DecrementRequest) (*countermapv1.DecrementResponse, error) {
	log.Debugw("Decrement",
		logging.Stringer("DecrementRequest", stringer.Truncate(request, truncLen)))
	input := &countermapv1.CounterMapInput{
		Input: &countermapv1.CounterMapInput_Decrement{
			Decrement: request.DecrementInput,
		},
	}
	output, headers, err := s.protocol.Command(ctx, input, request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Decrement",
			logging.Stringer("DecrementRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, err
	}
	response := &countermapv1.DecrementResponse{
		Headers:         headers,
		DecrementOutput: output.GetDecrement(),
	}
	log.Debugw("Decrement",
		logging.Stringer("DecrementRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("DecrementResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *CounterMapServer) Get(ctx context.Context, request *countermapv1.GetRequest) (*countermapv1.GetResponse, error) {
	log.Debugw("Get",
		logging.Stringer("GetRequest", stringer.Truncate(request, truncLen)))
	input := &countermapv1.CounterMapInput{
		Input: &countermapv1.CounterMapInput_Get{
			Get: request.GetInput,
		},
	}
	output, headers, err := s.protocol.Query(ctx, input, request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Get",
			logging.Stringer("GetRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, err
	}
	response := &countermapv1.GetResponse{
		Headers:   headers,
		GetOutput: output.GetGet(),
	}
	log.Debugw("Get",
		logging.Stringer("GetRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("GetResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *CounterMapServer) Remove(ctx context.Context, request *countermapv1.RemoveRequest) (*countermapv1.RemoveResponse, error) {
	log.Debugw("Remove",
		logging.Stringer("RemoveRequest", stringer.Truncate(request, truncLen)))
	input := &countermapv1.CounterMapInput{
		Input: &countermapv1.CounterMapInput_Remove{
			Remove: request.RemoveInput,
		},
	}
	output, headers, err := s.protocol.Command(ctx, input, request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Remove",
			logging.Stringer("RemoveRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, err
	}
	response := &countermapv1.RemoveResponse{
		Headers:      headers,
		RemoveOutput: output.GetRemove(),
	}
	log.Debugw("Remove",
		logging.Stringer("RemoveRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("RemoveResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *CounterMapServer) Clear(ctx context.Context, request *countermapv1.ClearRequest) (*countermapv1.ClearResponse, error) {
	log.Debugw("Clear",
		logging.Stringer("ClearRequest", stringer.Truncate(request, truncLen)))
	input := &countermapv1.CounterMapInput{
		Input: &countermapv1.CounterMapInput_Clear{
			Clear: request.ClearInput,
		},
	}
	output, headers, err := s.protocol.Command(ctx, input, request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Clear",
			logging.Stringer("ClearRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, err
	}
	response := &countermapv1.ClearResponse{
		Headers:     headers,
		ClearOutput: output.GetClear(),
	}
	log.Debugw("Clear",
		logging.Stringer("ClearRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("ClearResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *CounterMapServer) Lock(ctx context.Context, request *countermapv1.LockRequest) (*countermapv1.LockResponse, error) {
	log.Debugw("Lock",
		logging.Stringer("LockRequest", stringer.Truncate(request, truncLen)))
	input := &countermapv1.CounterMapInput{
		Input: &countermapv1.CounterMapInput_Lock{
			Lock: request.LockInput,
		},
	}
	output, headers, err := s.protocol.Command(ctx, input, request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Lock",
			logging.Stringer("LockRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, err
	}
	response := &countermapv1.LockResponse{
		Headers:    headers,
		LockOutput: output.GetLock(),
	}
	log.Debugw("Lock",
		logging.Stringer("LockRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("LockResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *CounterMapServer) Unlock(ctx context.Context, request *countermapv1.UnlockRequest) (*countermapv1.UnlockResponse, error) {
	log.Debugw("Unlock",
		logging.Stringer("UnlockRequest", stringer.Truncate(request, truncLen)))
	input := &countermapv1.CounterMapInput{
		Input: &countermapv1.CounterMapInput_Unlock{
			Unlock: request.UnlockInput,
		},
	}
	output, headers, err := s.protocol.Command(ctx, input, request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Unlock",
			logging.Stringer("UnlockRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, err
	}
	response := &countermapv1.UnlockResponse{
		Headers:      headers,
		UnlockOutput: output.GetUnlock(),
	}
	log.Debugw("Unlock",
		logging.Stringer("UnlockRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("UnlockResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *CounterMapServer) Events(request *countermapv1.EventsRequest, server countermapv1.CounterMap_EventsServer) error {
	log.Debugw("Events",
		logging.Stringer("EventsRequest", stringer.Truncate(request, truncLen)))
	input := &countermapv1.CounterMapInput{
		Input: &countermapv1.CounterMapInput_Events{
			Events: request.EventsInput,
		},
	}

	stream := streams.NewBufferedStream[*protocol.StreamCommandResponse[*countermapv1.CounterMapOutput]]()
	go func() {
		err := s.protocol.StreamCommand(server.Context(), input, request.Headers, stream)
		if err != nil {
			err = errors.ToProto(err)
			log.Warnw("Events",
				logging.Stringer("EventsRequest", stringer.Truncate(request, truncLen)),
				logging.Error("Error", err))
			stream.Error(err)
			stream.Close()
		}
	}()

	for {
		result, ok := stream.Receive()
		if !ok {
			return nil
		}

		if result.Failed() {
			err := errors.ToProto(result.Error)
			log.Warnw("Events",
				logging.Stringer("EventsRequest", stringer.Truncate(request, truncLen)),
				logging.Error("Error", err))
			return err
		}

		response := &countermapv1.EventsResponse{
			Headers:      result.Value.Headers,
			EventsOutput: result.Value.Output.GetEvents(),
		}
		log.Debugw("Events",
			logging.Stringer("EventsRequest", stringer.Truncate(request, truncLen)),
			logging.Stringer("EventsResponse", stringer.Truncate(response, truncLen)))
		if err := server.Send(response); err != nil {
			log.Warnw("Events",
				logging.Stringer("EventsRequest", stringer.Truncate(request, truncLen)),
				logging.Error("Error", err))
			return err
		}
	}
}

func (s *CounterMapServer) Entries(request *countermapv1.EntriesRequest, server countermapv1.CounterMap_EntriesServer) error {
	log.Debugw("Entries",
		logging.Stringer("EntriesRequest", stringer.Truncate(request, truncLen)))
	input := &countermapv1.CounterMapInput{
		Input: &countermapv1.CounterMapInput_Entries{
			Entries: request.EntriesInput,
		},
	}

	stream := streams.NewBufferedStream[*protocol.StreamQueryResponse[*countermapv1.CounterMapOutput]]()
	go func() {
		err := s.protocol.StreamQuery(server.Context(), input, request.Headers, stream)
		if err != nil {
			err = errors.ToProto(err)
			log.Warnw("Entries",
				logging.Stringer("EntriesRequest", stringer.Truncate(request, truncLen)),
				logging.Error("Error", err))
			stream.Error(err)
			stream.Close()
		}
	}()

	for {
		result, ok := stream.Receive()
		if !ok {
			return nil
		}

		if result.Failed() {
			err := errors.ToProto(result.Error)
			log.Warnw("Entries",
				logging.Stringer("EntriesRequest", stringer.Truncate(request, truncLen)),
				logging.Error("Error", err))
			return err
		}

		response := &countermapv1.EntriesResponse{
			Headers:       result.Value.Headers,
			EntriesOutput: result.Value.Output.GetEntries(),
		}
		log.Debugw("Entries",
			logging.Stringer("EntriesRequest", stringer.Truncate(request, truncLen)),
			logging.Stringer("EntriesResponse", stringer.Truncate(response, truncLen)))
		if err := server.Send(response); err != nil {
			log.Warnw("Entries",
				logging.Stringer("EntriesRequest", stringer.Truncate(request, truncLen)),
				logging.Error("Error", err))
			return err
		}
	}
}

var _ countermapv1.CounterMapServer = (*CounterMapServer)(nil)
