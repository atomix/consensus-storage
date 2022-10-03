// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"context"
	lockv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/lock/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/protocol"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/atomix/runtime/sdk/pkg/logging"
	"github.com/atomix/runtime/sdk/pkg/stringer"
	"github.com/gogo/protobuf/proto"
)

var log = logging.GetLogger()

const truncLen = 250

var lockCodec = protocol.NewCodec[*lockv1.LockInput, *lockv1.LockOutput](
	func(input *lockv1.LockInput) ([]byte, error) {
		return proto.Marshal(input)
	},
	func(bytes []byte) (*lockv1.LockOutput, error) {
		output := &lockv1.LockOutput{}
		if err := proto.Unmarshal(bytes, output); err != nil {
			return nil, err
		}
		return output, nil
	})

func NewLockServer(node *protocol.Node) lockv1.LockServer {
	return &LockServer{
		protocol: protocol.NewProtocol[*lockv1.LockInput, *lockv1.LockOutput](node, lockCodec),
	}
}

type LockServer struct {
	protocol protocol.Protocol[*lockv1.LockInput, *lockv1.LockOutput]
}

func (s *LockServer) Acquire(ctx context.Context, request *lockv1.AcquireRequest) (*lockv1.AcquireResponse, error) {
	log.Debugw("Acquire",
		logging.Stringer("AcquireRequest", stringer.Truncate(request, truncLen)))
	input := &lockv1.LockInput{
		Input: &lockv1.LockInput_Acquire{
			Acquire: request.AcquireInput,
		},
	}
	output, headers, err := s.protocol.Command(ctx, input, request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Acquire",
			logging.Stringer("AcquireRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, err
	}
	response := &lockv1.AcquireResponse{
		Headers:       headers,
		AcquireOutput: output.GetAcquire(),
	}
	log.Debugw("Acquire",
		logging.Stringer("AcquireRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("AcquireResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *LockServer) Release(ctx context.Context, request *lockv1.ReleaseRequest) (*lockv1.ReleaseResponse, error) {
	log.Debugw("Release",
		logging.Stringer("ReleaseRequest", stringer.Truncate(request, truncLen)))
	input := &lockv1.LockInput{
		Input: &lockv1.LockInput_Release{
			Release: request.ReleaseInput,
		},
	}
	output, headers, err := s.protocol.Command(ctx, input, request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Release",
			logging.Stringer("ReleaseRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, err
	}
	response := &lockv1.ReleaseResponse{
		Headers:       headers,
		ReleaseOutput: output.GetRelease(),
	}
	log.Debugw("Release",
		logging.Stringer("ReleaseRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("ReleaseResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *LockServer) Get(ctx context.Context, request *lockv1.GetRequest) (*lockv1.GetResponse, error) {
	log.Debugw("Get",
		logging.Stringer("GetRequest", stringer.Truncate(request, truncLen)))
	input := &lockv1.LockInput{
		Input: &lockv1.LockInput_Get{
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
	response := &lockv1.GetResponse{
		Headers:   headers,
		GetOutput: output.GetGet(),
	}
	log.Debugw("Get",
		logging.Stringer("GetRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("GetResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

var _ lockv1.LockServer = (*LockServer)(nil)
