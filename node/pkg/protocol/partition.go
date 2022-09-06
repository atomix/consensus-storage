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
	"github.com/lni/dragonboat/v3"
	"google.golang.org/grpc/metadata"
	"sync"
	"sync/atomic"
)

func newPartition(id multiraftv1.PartitionID, memberID multiraftv1.MemberID, host *dragonboat.NodeHost, streams *protocolContext) *Partition {
	return &Partition{
		id:       id,
		memberID: memberID,
		host:     host,
		streams:  streams,
	}
}

type Partition struct {
	id       multiraftv1.PartitionID
	memberID multiraftv1.MemberID
	host     *dragonboat.NodeHost
	ready    int32
	leader   uint64
	term     uint64
	streams  *protocolContext
	mu       sync.RWMutex
}

func (p *Partition) ID() multiraftv1.PartitionID {
	return p.id
}

func (p *Partition) setReady() {
	atomic.StoreInt32(&p.ready, 1)
}

func (p *Partition) getReady() bool {
	return atomic.LoadInt32(&p.ready) == 1
}

func (p *Partition) setLeader(term multiraftv1.Term, leader multiraftv1.MemberID) {
	atomic.StoreUint64(&p.term, uint64(term))
	atomic.StoreUint64(&p.leader, uint64(leader))
}

func (p *Partition) getLeader() (multiraftv1.Term, multiraftv1.MemberID) {
	return multiraftv1.Term(atomic.LoadUint64(&p.term)), multiraftv1.MemberID(atomic.LoadUint64(&p.leader))
}

func (p *Partition) Command(ctx context.Context, command *multiraftv1.StateMachineProposalInput) (*multiraftv1.StateMachineProposalOutput, error) {
	resultCh := make(chan streams.Result[*multiraftv1.StateMachineProposalOutput], 1)
	errCh := make(chan error, 1)
	go func() {
		if err := p.commitCommand(ctx, command, streams.NewChannelStream[*multiraftv1.StateMachineProposalOutput](resultCh)); err != nil {
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

		return result.Value, nil
	case err := <-errCh:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (p *Partition) StreamCommand(ctx context.Context, input *multiraftv1.StateMachineProposalInput, stream streams.WriteStream[*multiraftv1.StateMachineProposalOutput]) error {
	resultCh := make(chan streams.Result[*multiraftv1.StateMachineProposalOutput])
	go func() {
		if err := p.commitCommand(ctx, input, streams.NewChannelStream[*multiraftv1.StateMachineProposalOutput](resultCh)); err != nil {
			stream.Error(err)
			stream.Close()
			return
		}
	}()
	go func() {
		defer stream.Close()
		for result := range resultCh {
			stream.Send(result)
		}
	}()
	return nil
}

func (p *Partition) commitCommand(ctx context.Context, input *multiraftv1.StateMachineProposalInput, stream streams.WriteStream[*multiraftv1.StateMachineProposalOutput]) error {
	term, leader := p.getLeader()
	if leader != p.memberID {
		return errors.NewUnavailable("not the leader")
	}

	sequenceNum := p.streams.addStream(term, stream)
	proposal := &multiraftv1.RaftProposal{
		Term:        term,
		SequenceNum: sequenceNum,
		Proposal:    input,
	}

	bytes, err := proto.Marshal(proposal)
	if err != nil {
		return errors.NewInternal("failed to marshal RaftLogEntry: %v", err)
	}

	ctx, cancel := context.WithTimeout(ctx, defaultClientTimeout)
	defer cancel()
	if _, err := p.host.SyncPropose(ctx, p.host.GetNoOPSession(uint64(p.id)), bytes); err != nil {
		return wrapError(err)
	}
	return nil
}

func (p *Partition) Query(ctx context.Context, input *multiraftv1.StateMachineQueryInput) (*multiraftv1.StateMachineQueryOutput, error) {
	resultCh := make(chan streams.Result[*multiraftv1.StateMachineQueryOutput], 1)
	errCh := make(chan error, 1)
	query := &protocolQuery{
		input:  input,
		stream: streams.NewChannelStream[*multiraftv1.StateMachineQueryOutput](resultCh),
	}
	go func() {
		if err := p.applyQuery(ctx, query); err != nil {
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

		return result.Value, nil
	case err := <-errCh:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (p *Partition) StreamQuery(ctx context.Context, input *multiraftv1.StateMachineQueryInput, stream streams.WriteStream[*multiraftv1.StateMachineQueryOutput]) error {
	query := &protocolQuery{
		input:  input,
		stream: stream,
	}
	go func() {
		if err := p.applyQuery(ctx, query); err != nil {
			stream.Error(err)
			stream.Close()
			return
		}
	}()
	return nil
}

func (p *Partition) applyQuery(ctx context.Context, query *protocolQuery) error {
	md, _ := metadata.FromIncomingContext(ctx)
	sync := md["Sync"] != nil
	if sync {
		ctx, cancel := context.WithTimeout(ctx, defaultClientTimeout)
		defer cancel()
		if _, err := p.host.SyncRead(ctx, uint64(p.id), query); err != nil {
			return wrapError(err)
		}
	} else {
		if _, err := p.host.StaleRead(uint64(p.id), query); err != nil {
			return wrapError(err)
		}
	}
	return nil
}
