// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"context"
	countermapv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/atomic/countermap/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/protocol"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/atomix/runtime/sdk/pkg/logging"
	streams "github.com/atomix/runtime/sdk/pkg/stream"
	"github.com/gogo/protobuf/proto"
)

var log = logging.GetLogger()

var counterMapCodec = protocol.NewCodec[*countermapv1.AtomicCounterMapInput, *countermapv1.AtomicCounterMapOutput](
	func(input *countermapv1.AtomicCounterMapInput) ([]byte, error) {
		return proto.Marshal(input)
	},
	func(bytes []byte) (*countermapv1.AtomicCounterMapOutput, error) {
		output := &countermapv1.AtomicCounterMapOutput{}
		if err := proto.Unmarshal(bytes, output); err != nil {
			return nil, err
		}
		return output, nil
	})

func NewAtomicCounterMapServer(node *protocol.Node) countermapv1.AtomicCounterMapServer {
	return &AtomicCounterMapServer{
		protocol: protocol.NewProtocol[*countermapv1.AtomicCounterMapInput, *countermapv1.AtomicCounterMapOutput](node, counterMapCodec),
	}
}

type AtomicCounterMapServer struct {
	protocol protocol.Protocol[*countermapv1.AtomicCounterMapInput, *countermapv1.AtomicCounterMapOutput]
}

func (s *AtomicCounterMapServer) Size(ctx context.Context, request *countermapv1.SizeRequest) (*countermapv1.SizeResponse, error) {
	log.Debugw("Size",
		logging.Stringer("SizeRequest", request))
	input := &countermapv1.AtomicCounterMapInput{
		Input: &countermapv1.AtomicCounterMapInput_Size_{
			Size_: request.SizeInput,
		},
	}
	output, headers, err := s.protocol.Query(ctx, input, request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Size",
			logging.Stringer("SizeRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	response := &countermapv1.SizeResponse{
		Headers:    headers,
		SizeOutput: output.GetSize_(),
	}
	log.Debugw("Size",
		logging.Stringer("SizeRequest", request),
		logging.Stringer("SizeResponse", response))
	return response, nil
}

func (s *AtomicCounterMapServer) Set(ctx context.Context, request *countermapv1.SetRequest) (*countermapv1.SetResponse, error) {
	log.Debugw("Set",
		logging.Stringer("SetRequest", request))
	input := &countermapv1.AtomicCounterMapInput{
		Input: &countermapv1.AtomicCounterMapInput_Set{
			Set: request.SetInput,
		},
	}
	output, headers, err := s.protocol.Command(ctx, input, request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Set",
			logging.Stringer("SetRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	response := &countermapv1.SetResponse{
		Headers:   headers,
		SetOutput: output.GetSet(),
	}
	log.Debugw("Set",
		logging.Stringer("SetRequest", request),
		logging.Stringer("SetResponse", response))
	return response, nil
}

func (s *AtomicCounterMapServer) Insert(ctx context.Context, request *countermapv1.InsertRequest) (*countermapv1.InsertResponse, error) {
	log.Debugw("Insert",
		logging.Stringer("InsertRequest", request))
	input := &countermapv1.AtomicCounterMapInput{
		Input: &countermapv1.AtomicCounterMapInput_Insert{
			Insert: request.InsertInput,
		},
	}
	output, headers, err := s.protocol.Command(ctx, input, request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Insert",
			logging.Stringer("InsertRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	response := &countermapv1.InsertResponse{
		Headers:      headers,
		InsertOutput: output.GetInsert(),
	}
	log.Debugw("Insert",
		logging.Stringer("InsertRequest", request),
		logging.Stringer("InsertResponse", response))
	return response, nil
}

func (s *AtomicCounterMapServer) Update(ctx context.Context, request *countermapv1.UpdateRequest) (*countermapv1.UpdateResponse, error) {
	log.Debugw("Update",
		logging.Stringer("UpdateRequest", request))
	input := &countermapv1.AtomicCounterMapInput{
		Input: &countermapv1.AtomicCounterMapInput_Update{
			Update: request.UpdateInput,
		},
	}
	output, headers, err := s.protocol.Command(ctx, input, request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Update",
			logging.Stringer("UpdateRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	response := &countermapv1.UpdateResponse{
		Headers:      headers,
		UpdateOutput: output.GetUpdate(),
	}
	log.Debugw("Update",
		logging.Stringer("UpdateRequest", request),
		logging.Stringer("UpdateResponse", response))
	return response, nil
}

func (s *AtomicCounterMapServer) Increment(ctx context.Context, request *countermapv1.IncrementRequest) (*countermapv1.IncrementResponse, error) {
	log.Debugw("Increment",
		logging.Stringer("IncrementRequest", request))
	input := &countermapv1.AtomicCounterMapInput{
		Input: &countermapv1.AtomicCounterMapInput_Increment{
			Increment: request.IncrementInput,
		},
	}
	output, headers, err := s.protocol.Command(ctx, input, request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Increment",
			logging.Stringer("IncrementRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	response := &countermapv1.IncrementResponse{
		Headers:         headers,
		IncrementOutput: output.GetIncrement(),
	}
	log.Debugw("Increment",
		logging.Stringer("IncrementRequest", request),
		logging.Stringer("IncrementResponse", response))
	return response, nil
}

func (s *AtomicCounterMapServer) Decrement(ctx context.Context, request *countermapv1.DecrementRequest) (*countermapv1.DecrementResponse, error) {
	log.Debugw("Decrement",
		logging.Stringer("DecrementRequest", request))
	input := &countermapv1.AtomicCounterMapInput{
		Input: &countermapv1.AtomicCounterMapInput_Decrement{
			Decrement: request.DecrementInput,
		},
	}
	output, headers, err := s.protocol.Command(ctx, input, request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Decrement",
			logging.Stringer("DecrementRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	response := &countermapv1.DecrementResponse{
		Headers:         headers,
		DecrementOutput: output.GetDecrement(),
	}
	log.Debugw("Decrement",
		logging.Stringer("DecrementRequest", request),
		logging.Stringer("DecrementResponse", response))
	return response, nil
}

func (s *AtomicCounterMapServer) Get(ctx context.Context, request *countermapv1.GetRequest) (*countermapv1.GetResponse, error) {
	log.Debugw("Get",
		logging.Stringer("GetRequest", request))
	input := &countermapv1.AtomicCounterMapInput{
		Input: &countermapv1.AtomicCounterMapInput_Get{
			Get: request.GetInput,
		},
	}
	output, headers, err := s.protocol.Query(ctx, input, request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Get",
			logging.Stringer("GetRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	response := &countermapv1.GetResponse{
		Headers:   headers,
		GetOutput: output.GetGet(),
	}
	log.Debugw("Get",
		logging.Stringer("GetRequest", request),
		logging.Stringer("GetResponse", response))
	return response, nil
}

func (s *AtomicCounterMapServer) Remove(ctx context.Context, request *countermapv1.RemoveRequest) (*countermapv1.RemoveResponse, error) {
	log.Debugw("Remove",
		logging.Stringer("RemoveRequest", request))
	input := &countermapv1.AtomicCounterMapInput{
		Input: &countermapv1.AtomicCounterMapInput_Remove{
			Remove: request.RemoveInput,
		},
	}
	output, headers, err := s.protocol.Command(ctx, input, request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Remove",
			logging.Stringer("RemoveRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	response := &countermapv1.RemoveResponse{
		Headers:      headers,
		RemoveOutput: output.GetRemove(),
	}
	log.Debugw("Remove",
		logging.Stringer("RemoveRequest", request),
		logging.Stringer("RemoveResponse", response))
	return response, nil
}

func (s *AtomicCounterMapServer) Clear(ctx context.Context, request *countermapv1.ClearRequest) (*countermapv1.ClearResponse, error) {
	log.Debugw("Clear",
		logging.Stringer("ClearRequest", request))
	input := &countermapv1.AtomicCounterMapInput{
		Input: &countermapv1.AtomicCounterMapInput_Clear{
			Clear: request.ClearInput,
		},
	}
	output, headers, err := s.protocol.Command(ctx, input, request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Clear",
			logging.Stringer("ClearRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	response := &countermapv1.ClearResponse{
		Headers:     headers,
		ClearOutput: output.GetClear(),
	}
	log.Debugw("Clear",
		logging.Stringer("ClearRequest", request),
		logging.Stringer("ClearResponse", response))
	return response, nil
}

func (s *AtomicCounterMapServer) Lock(ctx context.Context, request *countermapv1.LockRequest) (*countermapv1.LockResponse, error) {
	log.Debugw("Lock",
		logging.Stringer("LockRequest", request))
	input := &countermapv1.AtomicCounterMapInput{
		Input: &countermapv1.AtomicCounterMapInput_Lock{
			Lock: request.LockInput,
		},
	}
	output, headers, err := s.protocol.Command(ctx, input, request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Lock",
			logging.Stringer("LockRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	response := &countermapv1.LockResponse{
		Headers:    headers,
		LockOutput: output.GetLock(),
	}
	log.Debugw("Lock",
		logging.Stringer("LockRequest", request),
		logging.Stringer("LockResponse", response))
	return response, nil
}

func (s *AtomicCounterMapServer) Unlock(ctx context.Context, request *countermapv1.UnlockRequest) (*countermapv1.UnlockResponse, error) {
	log.Debugw("Unlock",
		logging.Stringer("UnlockRequest", request))
	input := &countermapv1.AtomicCounterMapInput{
		Input: &countermapv1.AtomicCounterMapInput_Unlock{
			Unlock: request.UnlockInput,
		},
	}
	output, headers, err := s.protocol.Command(ctx, input, request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Unlock",
			logging.Stringer("UnlockRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	response := &countermapv1.UnlockResponse{
		Headers:      headers,
		UnlockOutput: output.GetUnlock(),
	}
	log.Debugw("Unlock",
		logging.Stringer("UnlockRequest", request),
		logging.Stringer("UnlockResponse", response))
	return response, nil
}

func (s *AtomicCounterMapServer) Events(request *countermapv1.EventsRequest, server countermapv1.AtomicCounterMap_EventsServer) error {
	log.Debugw("Events",
		logging.Stringer("EventsRequest", request))
	input := &countermapv1.AtomicCounterMapInput{
		Input: &countermapv1.AtomicCounterMapInput_Events{
			Events: request.EventsInput,
		},
	}

	ch := make(chan streams.Result[*protocol.StreamCommandResponse[*countermapv1.AtomicCounterMapOutput]])
	stream := streams.NewChannelStream[*protocol.StreamCommandResponse[*countermapv1.AtomicCounterMapOutput]](ch)
	go func() {
		err := s.protocol.StreamCommand(server.Context(), input, request.Headers, stream)
		if err != nil {
			err = errors.ToProto(err)
			log.Warnw("Events",
				logging.Stringer("EventsRequest", request),
				logging.Error("Error", err))
			stream.Error(err)
			stream.Close()
		}
	}()

	for result := range ch {
		if result.Failed() {
			err := errors.ToProto(result.Error)
			log.Warnw("Events",
				logging.Stringer("EventsRequest", request),
				logging.Error("Error", err))
			return err
		}

		response := &countermapv1.EventsResponse{
			Headers:      result.Value.Headers,
			EventsOutput: result.Value.Output.GetEvents(),
		}
		log.Debugw("Events",
			logging.Stringer("EventsRequest", request),
			logging.Stringer("EventsResponse", response))
		if err := server.Send(response); err != nil {
			log.Warnw("Events",
				logging.Stringer("EventsRequest", request),
				logging.Error("Error", err))
			return err
		}
	}
	return nil
}

func (s *AtomicCounterMapServer) Entries(request *countermapv1.EntriesRequest, server countermapv1.AtomicCounterMap_EntriesServer) error {
	log.Debugw("Entries",
		logging.Stringer("EntriesRequest", request))
	input := &countermapv1.AtomicCounterMapInput{
		Input: &countermapv1.AtomicCounterMapInput_Entries{
			Entries: request.EntriesInput,
		},
	}

	ch := make(chan streams.Result[*protocol.StreamQueryResponse[*countermapv1.AtomicCounterMapOutput]])
	stream := streams.NewChannelStream[*protocol.StreamQueryResponse[*countermapv1.AtomicCounterMapOutput]](ch)
	go func() {
		err := s.protocol.StreamQuery(server.Context(), input, request.Headers, stream)
		if err != nil {
			err = errors.ToProto(err)
			log.Warnw("Entries",
				logging.Stringer("EntriesRequest", request),
				logging.Error("Error", err))
			stream.Error(err)
			stream.Close()
		}
	}()

	for result := range ch {
		if result.Failed() {
			err := errors.ToProto(result.Error)
			log.Warnw("Entries",
				logging.Stringer("EntriesRequest", request),
				logging.Error("Error", err))
			return err
		}

		response := &countermapv1.EntriesResponse{
			Headers:       result.Value.Headers,
			EntriesOutput: result.Value.Output.GetEntries(),
		}
		log.Debugw("Entries",
			logging.Stringer("EntriesRequest", request),
			logging.Stringer("EntriesResponse", response))
		if err := server.Send(response); err != nil {
			log.Warnw("Entries",
				logging.Stringer("EntriesRequest", request),
				logging.Error("Error", err))
			return err
		}
	}
	return nil
}

var _ countermapv1.AtomicCounterMapServer = (*AtomicCounterMapServer)(nil)
