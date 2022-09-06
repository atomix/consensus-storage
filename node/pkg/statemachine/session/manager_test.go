// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine"
	"github.com/atomix/runtime/sdk/pkg/logging"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestManager(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	parentCtx := statemachine.NewMockSessionManagerContext(ctrl)
	parentCtx.EXPECT().Time().Return(time.Now()).AnyTimes()
	parentCtx.EXPECT().Log().Return(logging.GetLogger()).AnyTimes()

	primitives := NewMockPrimitiveManager(ctrl)
	manager := NewManager(parentCtx, func(smCtx Context) PrimitiveManager {
		return primitives
	})
	context := manager.(Context)

	// Open a new session
	parentCtx.EXPECT().Index().Return(statemachine.Index(1)).AnyTimes()
	openSession := statemachine.NewMockOpenSessionProposal(ctrl)
	openSession.EXPECT().ID().Return(statemachine.ProposalID(1)).AnyTimes()
	openSession.EXPECT().Input().Return(&multiraftv1.OpenSessionInput{
		Timeout: time.Minute,
	}).AnyTimes()
	openSession.EXPECT().Close()
	var sessionID *multiraftv1.SessionID
	openSession.EXPECT().Output(gomock.Any()).Times(1).Do(func(output *multiraftv1.OpenSessionOutput) {
		sessionID = &output.SessionID
	})
	manager.OpenSession(openSession)
	assert.NotNil(t, sessionID)
	assert.Len(t, context.Sessions().List(), 1)
	assert.Equal(t, ID(*sessionID), context.Sessions().List()[0].ID())

	// Create a primitive using the session
	parentCtx.EXPECT().Index().Return(statemachine.Index(2)).AnyTimes()
	proposal := statemachine.NewMockSessionProposal(ctrl)
	proposal.EXPECT().ID().Return(statemachine.ProposalID(2)).AnyTimes()
	proposal.EXPECT().Input().Return(&multiraftv1.SessionProposalInput{
		SessionID:   *sessionID,
		SequenceNum: 1,
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
		assert.Equal(t, ID(*sessionID), proposal.Session().ID())
		assert.Len(t, proposal.Session().Proposals().List(), 0)
		assert.Len(t, context.Proposals().List(), 0)
		proposal.Output(&multiraftv1.CreatePrimitiveOutput{
			PrimitiveID: 1,
		})
		proposal.Close()
	})
	manager.Propose(proposal)

	// Submit a primitive proposal and verify the proposal is applied to the primitive
	parentCtx.EXPECT().Index().Return(statemachine.Index(3)).AnyTimes()
	proposal = statemachine.NewMockSessionProposal(ctrl)
	proposal.EXPECT().ID().Return(statemachine.ProposalID(3)).AnyTimes()
	proposal.EXPECT().Input().Return(&multiraftv1.SessionProposalInput{
		SessionID:   *sessionID,
		SequenceNum: 2,
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
		assert.Equal(t, ID(*sessionID), proposal.Session().ID())
		assert.Len(t, proposal.Session().Proposals().List(), 1)
		assert.Len(t, context.Proposals().List(), 1)
		proposal.Output(&multiraftv1.PrimitiveProposalOutput{
			Payload: []byte("world!"),
		})
		proposal.Close()
		assert.Len(t, proposal.Session().Proposals().List(), 0)
		assert.Len(t, context.Proposals().List(), 0)
	})
	manager.Propose(proposal)

	// Retry the same primitive proposal and verify the proposal is not applied to the primitive again (for linearizability)
	parentCtx.EXPECT().Index().Return(statemachine.Index(4)).AnyTimes()
	proposal = statemachine.NewMockSessionProposal(ctrl)
	proposal.EXPECT().ID().Return(statemachine.ProposalID(4)).AnyTimes()
	proposal.EXPECT().Input().Return(&multiraftv1.SessionProposalInput{
		SessionID:   *sessionID,
		SequenceNum: 2,
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
	query.EXPECT().ID().Return(statemachine.QueryID(1)).AnyTimes()
	query.EXPECT().Input().Return(&multiraftv1.SessionQueryInput{
		SessionID: *sessionID,
		Input: &multiraftv1.SessionQueryInput_Query{
			Query: &multiraftv1.PrimitiveQueryInput{
				PrimitiveID: 1,
			},
		},
	}).AnyTimes()
	query.EXPECT().Output(gomock.Any())
	query.EXPECT().Close()
	primitives.EXPECT().Query(gomock.Any()).Do(func(query PrimitiveQuery) {
		assert.Equal(t, ID(*sessionID), query.Session().ID())
		query.Output(&multiraftv1.PrimitiveQueryOutput{})
		query.Close()
	})
	manager.Query(query)

	// Close the primitive
	parentCtx.EXPECT().Index().Return(statemachine.Index(5)).AnyTimes()
	proposal = statemachine.NewMockSessionProposal(ctrl)
	proposal.EXPECT().ID().Return(statemachine.ProposalID(5)).AnyTimes()
	proposal.EXPECT().Input().Return(&multiraftv1.SessionProposalInput{
		SessionID:   *sessionID,
		SequenceNum: 3,
		Input: &multiraftv1.SessionProposalInput_ClosePrimitive{
			ClosePrimitive: &multiraftv1.ClosePrimitiveInput{
				PrimitiveID: 1,
			},
		},
	}).AnyTimes()
	proposal.EXPECT().Output(gomock.Any())
	proposal.EXPECT().Close()
	primitives.EXPECT().ClosePrimitive(gomock.Any()).Do(func(proposal ClosePrimitiveProposal) {
		assert.Equal(t, ID(*sessionID), proposal.Session().ID())
		assert.Len(t, proposal.Session().Proposals().List(), 0)
		assert.Len(t, context.Proposals().List(), 0)
		proposal.Output(&multiraftv1.ClosePrimitiveOutput{})
		proposal.Close()
	})
	manager.Propose(proposal)

	// Close the session and verify it is removed from the session manager context
	parentCtx.EXPECT().Index().Return(statemachine.Index(6)).AnyTimes()
	closeSession := statemachine.NewMockCloseSessionProposal(ctrl)
	closeSession.EXPECT().ID().Return(statemachine.ProposalID(6)).AnyTimes()
	closeSession.EXPECT().Input().Return(&multiraftv1.CloseSessionInput{
		SessionID: *sessionID,
	}).AnyTimes()
	closeSession.EXPECT().Output(gomock.Any())
	closeSession.EXPECT().Close()
	manager.CloseSession(closeSession)
	assert.Len(t, context.Sessions().List(), 0)
}
