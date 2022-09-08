// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package primitive

import (
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/session"
	"github.com/atomix/runtime/sdk/pkg/logging"
	"github.com/golang/mock/gomock"
	"testing"
	"time"
)

func newManagerTestContext(ctrl *gomock.Controller) *managerTestContext {
	context := session.NewMockContext(ctrl)
	context.EXPECT().Time().Return(time.UnixMilli(0)).AnyTimes()
	context.EXPECT().Log().Return(logging.GetLogger()).AnyTimes()
	return &managerTestContext{
		MockContext: context,
	}
}

type managerTestContext struct {
	*session.MockContext
	index   statemachine.Index
	queryID statemachine.QueryID
}

func (c *managerTestContext) nextProposalID() statemachine.ProposalID {
	c.index++
	c.EXPECT().Index().Return(c.index).AnyTimes()
	return statemachine.ProposalID(c.index)
}

func (c *managerTestContext) nextQueryID() statemachine.QueryID {
	c.queryID++
	return c.queryID
}

func TestPrimitive(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	context := newManagerTestContext(ctrl)
	timer := statemachine.NewMockTimer(ctrl)
	scheduler := statemachine.NewMockScheduler(ctrl)
	scheduler.EXPECT().Schedule(gomock.Any(), gomock.Any()).Return(timer).AnyTimes()
	context.EXPECT().Scheduler().Return(scheduler).AnyTimes()

	registry := NewTypeRegistry()
	codec := NewMockAnyCodec(ctrl)
	codec.EXPECT().DecodeInput(gomock.Any()).Return("Hello", nil).AnyTimes()
	codec.EXPECT().EncodeOutput(gomock.Any()).Return([]byte("world!"), nil).AnyTimes()
	primitive := NewMockAnyPrimitive(ctrl)
	primitiveType := NewType[any, any]("test", codec, func(Context[any, any]) Primitive[any, any] {
		return primitive
	})
	RegisterType[any, any](registry)(primitiveType)

	manager := NewManager(context, registry)

	primitiveSession := session.NewMockSession(ctrl)
	primitiveSession.EXPECT().ID().Return(session.ID(1)).AnyTimes()
	primitiveSession.EXPECT().Log().Return(logging.GetLogger()).AnyTimes()
	primitiveSession.EXPECT().Watch(gomock.Any()).Return(func() {}).AnyTimes()

	createPrimitive := session.NewMockCreatePrimitiveProposal(ctrl)
	createPrimitive.EXPECT().ID().Return(context.nextProposalID()).AnyTimes()
	createPrimitive.EXPECT().Session().Return(primitiveSession).AnyTimes()
	createPrimitive.EXPECT().Input().Return(&multiraftv1.CreatePrimitiveInput{
		PrimitiveSpec: multiraftv1.PrimitiveSpec{
			Service:   "test",
			Namespace: "test-namespace",
			Name:      "test-name",
		},
	}).AnyTimes()
	var primitiveID multiraftv1.PrimitiveID
	createPrimitive.EXPECT().Output(gomock.Any()).Do(func(output *multiraftv1.CreatePrimitiveOutput) {
		primitiveID = output.PrimitiveID
	})
	createPrimitive.EXPECT().Close()
	manager.CreatePrimitive(createPrimitive)

	proposal := session.NewMockPrimitiveProposal(ctrl)
	proposal.EXPECT().ID().Return(context.nextProposalID()).AnyTimes()
	proposal.EXPECT().Session().Return(primitiveSession).AnyTimes()
	proposal.EXPECT().Input().Return(&multiraftv1.PrimitiveProposalInput{
		PrimitiveID: primitiveID,
		Payload:     []byte("Hello"),
	}).AnyTimes()
	proposal.EXPECT().Output(gomock.Any())
	proposal.EXPECT().Close()
	primitive.EXPECT().Propose(gomock.Any()).Do(func(proposal Proposal[any, any]) {
		proposal.Output("world!")
		proposal.Close()
	})
	manager.Propose(proposal)

	query := session.NewMockPrimitiveQuery(ctrl)
	query.EXPECT().ID().Return(context.nextQueryID()).AnyTimes()
	query.EXPECT().Session().Return(primitiveSession).AnyTimes()
	query.EXPECT().Input().Return(&multiraftv1.PrimitiveQueryInput{
		PrimitiveID: primitiveID,
		Payload:     []byte("Hello"),
	}).AnyTimes()
	query.EXPECT().Output(gomock.Any())
	query.EXPECT().Close()
	primitive.EXPECT().Query(gomock.Any()).Do(func(query Query[any, any]) {
		query.Output("world!")
		query.Close()
	})
	manager.Query(query)

	closePrimitive := session.NewMockClosePrimitiveProposal(ctrl)
	closePrimitive.EXPECT().ID().Return(context.nextProposalID()).AnyTimes()
	closePrimitive.EXPECT().Session().Return(primitiveSession).AnyTimes()
	closePrimitive.EXPECT().Input().Return(&multiraftv1.ClosePrimitiveInput{
		PrimitiveID: primitiveID,
	}).AnyTimes()
	closePrimitive.EXPECT().Output(gomock.Any())
	closePrimitive.EXPECT().Close()
	manager.ClosePrimitive(closePrimitive)
}
