// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"bytes"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/snapshot"
	"github.com/atomix/runtime/sdk/pkg/logging"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func newManagerTestContext(ctrl *gomock.Controller) *managerTestContext {
	context := statemachine.NewMockSessionManagerContext(ctrl)
	context.EXPECT().Time().Return(time.Now()).AnyTimes()
	context.EXPECT().Log().Return(logging.GetLogger()).AnyTimes()
	return &managerTestContext{
		MockSessionManagerContext: context,
		sequenceNums:              make(map[ID]multiraftv1.SequenceNum),
	}
}

type managerTestContext struct {
	*statemachine.MockSessionManagerContext
	index        statemachine.Index
	sequenceNums map[ID]multiraftv1.SequenceNum
	queryID      statemachine.QueryID
}

func (c *managerTestContext) nextProposalID() statemachine.ProposalID {
	c.index++
	c.EXPECT().Index().Return(c.index).AnyTimes()
	return statemachine.ProposalID(c.index)
}

func (c *managerTestContext) nextSequenceNum(sessionID ID) multiraftv1.SequenceNum {
	sequenceNum := c.sequenceNums[sessionID] + 1
	c.sequenceNums[sessionID] = sequenceNum
	return sequenceNum
}

func (c *managerTestContext) nextQueryID() statemachine.QueryID {
	c.queryID++
	return c.queryID
}

func TestManager(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	context := newManagerTestContext(ctrl)

	primitives := NewMockPrimitiveManager(ctrl)
	manager := NewManager(context, func(Context) PrimitiveManager {
		return primitives
	})

	// Open a new session
	proposalID := context.nextProposalID()
	sessionID := ID(proposalID)
	openSession := statemachine.NewMockOpenSessionProposal(ctrl)
	openSession.EXPECT().ID().Return(proposalID).AnyTimes()
	openSession.EXPECT().Input().Return(&multiraftv1.OpenSessionInput{
		Timeout: time.Minute,
	}).AnyTimes()
	openSession.EXPECT().Close()
	openSession.EXPECT().Output(gomock.Any())
	manager.OpenSession(openSession)
	assert.Len(t, manager.(Context).Sessions().List(), 1)
	assert.Equal(t, sessionID, manager.(Context).Sessions().List()[0].ID())

	// Take a snapshot of the manager and create a new manager from the snapshot
	buf := &bytes.Buffer{}
	primitives.EXPECT().Snapshot(gomock.Any()).Return(nil)
	assert.NoError(t, manager.Snapshot(snapshot.NewWriter(buf)))
	manager = NewManager(context, func(Context) PrimitiveManager {
		return primitives
	})
	primitives.EXPECT().Recover(gomock.Any()).Return(nil)
	assert.NoError(t, manager.Recover(snapshot.NewReader(buf)))
	assert.Len(t, manager.(Context).Sessions().List(), 1)
	assert.Equal(t, sessionID, manager.(Context).Sessions().List()[0].ID())

	// Create a primitive using the session
	proposal := statemachine.NewMockSessionProposal(ctrl)
	proposalID = context.nextProposalID()
	sequenceNum := context.nextSequenceNum(sessionID)
	proposal.EXPECT().ID().Return(proposalID).AnyTimes()
	proposal.EXPECT().Input().Return(&multiraftv1.SessionProposalInput{
		SessionID:   multiraftv1.SessionID(sessionID),
		SequenceNum: sequenceNum,
		Input: &multiraftv1.SessionProposalInput_CreatePrimitive{
			CreatePrimitive: &multiraftv1.CreatePrimitiveInput{
				PrimitiveSpec: multiraftv1.PrimitiveSpec{
					Service:   "test",
					Namespace: "foo",
					Name:      "bar",
				},
			},
		},
	}).AnyTimes()
	proposal.EXPECT().Output(gomock.Any())
	proposal.EXPECT().Close()
	primitives.EXPECT().CreatePrimitive(gomock.Any()).Do(func(proposal CreatePrimitiveProposal) {
		assert.Equal(t, sessionID, proposal.Session().ID())
		assert.Len(t, proposal.Session().Proposals().List(), 0)
		assert.Len(t, manager.(Context).Proposals().List(), 0)
		proposal.Output(&multiraftv1.CreatePrimitiveOutput{
			PrimitiveID: 1,
		})
		proposal.Close()
	})
	manager.Propose(proposal)

	// Submit a primitive proposal and verify the proposal is applied to the primitive
	proposal = statemachine.NewMockSessionProposal(ctrl)
	proposalID = context.nextProposalID()
	sequenceNum = context.nextSequenceNum(sessionID)
	proposal.EXPECT().ID().Return(proposalID).AnyTimes()
	proposal.EXPECT().Input().Return(&multiraftv1.SessionProposalInput{
		SessionID:   multiraftv1.SessionID(sessionID),
		SequenceNum: sequenceNum,
		Input: &multiraftv1.SessionProposalInput_Proposal{
			Proposal: &multiraftv1.PrimitiveProposalInput{
				PrimitiveID: 1,
				Payload:     []byte("Hello"),
			},
		},
	}).AnyTimes()
	proposal.EXPECT().Output(gomock.Any())
	proposal.EXPECT().Close()
	primitives.EXPECT().Propose(gomock.Any()).Do(func(proposal PrimitiveProposal) {
		assert.Equal(t, proposalID, proposal.ID())
		assert.Equal(t, sessionID, proposal.Session().ID())
		assert.Len(t, proposal.Session().Proposals().List(), 1)
		p, ok := proposal.Session().Proposals().Get(3)
		assert.True(t, ok)
		assert.Equal(t, proposalID, p.ID())
		assert.Len(t, manager.(Context).Proposals().List(), 1)
		p, ok = manager.(Context).Proposals().Get(3)
		assert.True(t, ok)
		assert.Equal(t, proposalID, p.ID())
		proposal.Output(&multiraftv1.PrimitiveProposalOutput{
			Payload: []byte("world!"),
		})
		proposal.Close()
		assert.Len(t, proposal.Session().Proposals().List(), 0)
		assert.Len(t, manager.(Context).Proposals().List(), 0)
	})
	manager.Propose(proposal)

	// Retry the same primitive proposal and verify the proposal is not applied to the primitive again (for linearizability)
	proposal = statemachine.NewMockSessionProposal(ctrl)
	proposalID = context.nextProposalID()
	proposal.EXPECT().ID().Return(proposalID).AnyTimes()
	proposal.EXPECT().Input().Return(&multiraftv1.SessionProposalInput{
		SessionID:   multiraftv1.SessionID(sessionID),
		SequenceNum: sequenceNum,
		Input: &multiraftv1.SessionProposalInput_Proposal{
			Proposal: &multiraftv1.PrimitiveProposalInput{
				PrimitiveID: 1,
				Payload:     []byte("Hello"),
			},
		},
	}).AnyTimes()
	proposal.EXPECT().Output(gomock.Any())
	proposal.EXPECT().Close()
	manager.Propose(proposal)

	// Take another snapshot of the manager and create a new manager from the snapshot
	assert.Len(t, manager.(Context).Proposals().List(), 0)
	assert.Len(t, manager.(Context).Sessions().List()[0].Proposals().List(), 0)
	buf = &bytes.Buffer{}
	primitives.EXPECT().Snapshot(gomock.Any()).Return(nil)
	assert.NoError(t, manager.Snapshot(snapshot.NewWriter(buf)))
	manager = NewManager(context, func(Context) PrimitiveManager {
		return primitives
	})
	primitives.EXPECT().Recover(gomock.Any()).Return(nil)
	assert.NoError(t, manager.Recover(snapshot.NewReader(buf)))
	assert.Len(t, manager.(Context).Sessions().List(), 1)
	assert.Equal(t, sessionID, manager.(Context).Sessions().List()[0].ID())
	assert.Len(t, manager.(Context).Proposals().List(), 0)
	assert.Len(t, manager.(Context).Sessions().List()[0].Proposals().List(), 0)

	// Retry the same primitive proposal again after the snapshot
	proposal = statemachine.NewMockSessionProposal(ctrl)
	proposalID = context.nextProposalID()
	proposal.EXPECT().ID().Return(proposalID).AnyTimes()
	proposal.EXPECT().Input().Return(&multiraftv1.SessionProposalInput{
		SessionID:   multiraftv1.SessionID(sessionID),
		SequenceNum: sequenceNum,
		Input: &multiraftv1.SessionProposalInput_Proposal{
			Proposal: &multiraftv1.PrimitiveProposalInput{
				PrimitiveID: 1,
				Payload:     []byte("Hello"),
			},
		},
	}).AnyTimes()
	proposal.EXPECT().Output(gomock.Any())
	proposal.EXPECT().Close()
	manager.Propose(proposal)

	// Submit a primitive query and verify the query is applied to the primitive
	query := statemachine.NewMockSessionQuery(ctrl)
	queryID := context.nextQueryID()
	query.EXPECT().ID().Return(queryID).AnyTimes()
	query.EXPECT().Input().Return(&multiraftv1.SessionQueryInput{
		SessionID: multiraftv1.SessionID(sessionID),
		Input: &multiraftv1.SessionQueryInput_Query{
			Query: &multiraftv1.PrimitiveQueryInput{
				PrimitiveID: 1,
			},
		},
	}).AnyTimes()
	query.EXPECT().Output(gomock.Any())
	query.EXPECT().Close()
	primitives.EXPECT().Query(gomock.Any()).Do(func(query PrimitiveQuery) {
		assert.Equal(t, queryID, query.ID())
		assert.Equal(t, sessionID, query.Session().ID())
		query.Output(&multiraftv1.PrimitiveQueryOutput{})
		query.Close()
	})
	manager.Query(query)

	// Close the primitive
	proposal = statemachine.NewMockSessionProposal(ctrl)
	proposalID = context.nextProposalID()
	sequenceNum = context.nextSequenceNum(sessionID)
	proposal.EXPECT().ID().Return(proposalID).AnyTimes()
	proposal.EXPECT().Input().Return(&multiraftv1.SessionProposalInput{
		SessionID:   multiraftv1.SessionID(sessionID),
		SequenceNum: sequenceNum,
		Input: &multiraftv1.SessionProposalInput_ClosePrimitive{
			ClosePrimitive: &multiraftv1.ClosePrimitiveInput{
				PrimitiveID: 1,
			},
		},
	}).AnyTimes()
	proposal.EXPECT().Output(gomock.Any())
	proposal.EXPECT().Close()
	primitives.EXPECT().ClosePrimitive(gomock.Any()).Do(func(proposal ClosePrimitiveProposal) {
		assert.Equal(t, sessionID, proposal.Session().ID())
		assert.Len(t, proposal.Session().Proposals().List(), 0)
		assert.Len(t, manager.(Context).Proposals().List(), 0)
		proposal.Output(&multiraftv1.ClosePrimitiveOutput{})
		proposal.Close()
	})
	manager.Propose(proposal)

	// Close the session and verify it is removed from the session manager context
	closeSession := statemachine.NewMockCloseSessionProposal(ctrl)
	proposalID = context.nextProposalID()
	closeSession.EXPECT().ID().Return(proposalID).AnyTimes()
	closeSession.EXPECT().Input().Return(&multiraftv1.CloseSessionInput{
		SessionID: multiraftv1.SessionID(sessionID),
	}).AnyTimes()
	closeSession.EXPECT().Output(gomock.Any())
	closeSession.EXPECT().Close()
	manager.CloseSession(closeSession)
	assert.Len(t, manager.(Context).Sessions().List(), 0)
}
