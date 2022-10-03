// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"context"
	electionv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/election/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/protocol"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/atomix/runtime/sdk/pkg/logging"
	streams "github.com/atomix/runtime/sdk/pkg/stream"
	"github.com/atomix/runtime/sdk/pkg/stringer"
	"github.com/gogo/protobuf/proto"
)

var log = logging.GetLogger()

const truncLen = 250

var lockCodec = protocol.NewCodec[*electionv1.LeaderElectionInput, *electionv1.LeaderElectionOutput](
	func(input *electionv1.LeaderElectionInput) ([]byte, error) {
		return proto.Marshal(input)
	},
	func(bytes []byte) (*electionv1.LeaderElectionOutput, error) {
		output := &electionv1.LeaderElectionOutput{}
		if err := proto.Unmarshal(bytes, output); err != nil {
			return nil, err
		}
		return output, nil
	})

func NewLeaderElectionServer(node *protocol.Node) electionv1.LeaderElectionServer {
	return &LeaderElectionServer{
		protocol: protocol.NewProtocol[*electionv1.LeaderElectionInput, *electionv1.LeaderElectionOutput](node, lockCodec),
	}
}

type LeaderElectionServer struct {
	protocol protocol.Protocol[*electionv1.LeaderElectionInput, *electionv1.LeaderElectionOutput]
}

func (s *LeaderElectionServer) Enter(ctx context.Context, request *electionv1.EnterRequest) (*electionv1.EnterResponse, error) {
	log.Debugw("Enter",
		logging.Stringer("EnterRequest", stringer.Truncate(request, truncLen)))
	input := &electionv1.LeaderElectionInput{
		Input: &electionv1.LeaderElectionInput_Enter{
			Enter: request.EnterInput,
		},
	}
	output, headers, err := s.protocol.Command(ctx, input, request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Enter",
			logging.Stringer("EnterRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, err
	}
	response := &electionv1.EnterResponse{
		Headers:     headers,
		EnterOutput: output.GetEnter(),
	}
	log.Debugw("Enter",
		logging.Stringer("EnterRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("EnterResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *LeaderElectionServer) Withdraw(ctx context.Context, request *electionv1.WithdrawRequest) (*electionv1.WithdrawResponse, error) {
	log.Debugw("Withdraw",
		logging.Stringer("WithdrawRequest", stringer.Truncate(request, truncLen)))
	input := &electionv1.LeaderElectionInput{
		Input: &electionv1.LeaderElectionInput_Withdraw{
			Withdraw: request.WithdrawInput,
		},
	}
	output, headers, err := s.protocol.Command(ctx, input, request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Withdraw",
			logging.Stringer("WithdrawRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, err
	}
	response := &electionv1.WithdrawResponse{
		Headers:        headers,
		WithdrawOutput: output.GetWithdraw(),
	}
	log.Debugw("Withdraw",
		logging.Stringer("WithdrawRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("WithdrawResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *LeaderElectionServer) Anoint(ctx context.Context, request *electionv1.AnointRequest) (*electionv1.AnointResponse, error) {
	log.Debugw("Anoint",
		logging.Stringer("AnointRequest", stringer.Truncate(request, truncLen)))
	input := &electionv1.LeaderElectionInput{
		Input: &electionv1.LeaderElectionInput_Anoint{
			Anoint: request.AnointInput,
		},
	}
	output, headers, err := s.protocol.Command(ctx, input, request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Anoint",
			logging.Stringer("AnointRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, err
	}
	response := &electionv1.AnointResponse{
		Headers:      headers,
		AnointOutput: output.GetAnoint(),
	}
	log.Debugw("Anoint",
		logging.Stringer("AnointRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("AnointResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *LeaderElectionServer) Promote(ctx context.Context, request *electionv1.PromoteRequest) (*electionv1.PromoteResponse, error) {
	log.Debugw("Promote",
		logging.Stringer("PromoteRequest", stringer.Truncate(request, truncLen)))
	input := &electionv1.LeaderElectionInput{
		Input: &electionv1.LeaderElectionInput_Promote{
			Promote: request.PromoteInput,
		},
	}
	output, headers, err := s.protocol.Command(ctx, input, request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Promote",
			logging.Stringer("PromoteRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, err
	}
	response := &electionv1.PromoteResponse{
		Headers:       headers,
		PromoteOutput: output.GetPromote(),
	}
	log.Debugw("Promote",
		logging.Stringer("PromoteRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("PromoteResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *LeaderElectionServer) Demote(ctx context.Context, request *electionv1.DemoteRequest) (*electionv1.DemoteResponse, error) {
	log.Debugw("Demote",
		logging.Stringer("DemoteRequest", stringer.Truncate(request, truncLen)))
	input := &electionv1.LeaderElectionInput{
		Input: &electionv1.LeaderElectionInput_Demote{
			Demote: request.DemoteInput,
		},
	}
	output, headers, err := s.protocol.Command(ctx, input, request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Demote",
			logging.Stringer("DemoteRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, err
	}
	response := &electionv1.DemoteResponse{
		Headers:      headers,
		DemoteOutput: output.GetDemote(),
	}
	log.Debugw("Demote",
		logging.Stringer("DemoteRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("DemoteResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *LeaderElectionServer) Evict(ctx context.Context, request *electionv1.EvictRequest) (*electionv1.EvictResponse, error) {
	log.Debugw("Evict",
		logging.Stringer("EvictRequest", stringer.Truncate(request, truncLen)))
	input := &electionv1.LeaderElectionInput{
		Input: &electionv1.LeaderElectionInput_Evict{
			Evict: request.EvictInput,
		},
	}
	output, headers, err := s.protocol.Command(ctx, input, request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Evict",
			logging.Stringer("EvictRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, err
	}
	response := &electionv1.EvictResponse{
		Headers:     headers,
		EvictOutput: output.GetEvict(),
	}
	log.Debugw("Evict",
		logging.Stringer("EvictRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("EvictResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *LeaderElectionServer) GetTerm(ctx context.Context, request *electionv1.GetTermRequest) (*electionv1.GetTermResponse, error) {
	log.Debugw("GetTerm",
		logging.Stringer("GetTermRequest", stringer.Truncate(request, truncLen)))
	input := &electionv1.LeaderElectionInput{
		Input: &electionv1.LeaderElectionInput_GetTerm{
			GetTerm: request.GetTermInput,
		},
	}
	output, headers, err := s.protocol.Query(ctx, input, request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("GetTerm",
			logging.Stringer("GetTermRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, err
	}
	response := &electionv1.GetTermResponse{
		Headers:       headers,
		GetTermOutput: output.GetGetTerm(),
	}
	log.Debugw("GetTerm",
		logging.Stringer("GetTermRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("GetTermResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *LeaderElectionServer) Watch(request *electionv1.WatchRequest, server electionv1.LeaderElection_WatchServer) error {
	log.Debugw("Watch",
		logging.Stringer("WatchRequest", stringer.Truncate(request, truncLen)))
	input := &electionv1.LeaderElectionInput{
		Input: &electionv1.LeaderElectionInput_Watch{
			Watch: request.WatchInput,
		},
	}

	stream := streams.NewBufferedStream[*protocol.StreamQueryResponse[*electionv1.LeaderElectionOutput]]()
	go func() {
		err := s.protocol.StreamQuery(server.Context(), input, request.Headers, stream)
		if err != nil {
			err = errors.ToProto(err)
			log.Warnw("Watch",
				logging.Stringer("WatchRequest", stringer.Truncate(request, truncLen)),
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
			log.Warnw("Watch",
				logging.Stringer("WatchRequest", stringer.Truncate(request, truncLen)),
				logging.Error("Error", err))
			return err
		}

		response := &electionv1.WatchResponse{
			Headers:     result.Value.Headers,
			WatchOutput: result.Value.Output.GetWatch(),
		}
		log.Debugw("Watch",
			logging.Stringer("WatchRequest", stringer.Truncate(request, truncLen)),
			logging.Stringer("WatchResponse", stringer.Truncate(response, truncLen)))
		if err := server.Send(response); err != nil {
			log.Warnw("Watch",
				logging.Stringer("WatchRequest", stringer.Truncate(request, truncLen)),
				logging.Error("Error", err))
			return err
		}
	}
}

var _ electionv1.LeaderElectionServer = (*LeaderElectionServer)(nil)
