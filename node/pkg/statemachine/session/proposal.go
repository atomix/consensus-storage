// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"container/list"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/snapshot"
	"github.com/atomix/runtime/sdk/pkg/logging"
	"github.com/google/uuid"
	"time"
)

type ProposalID = statemachine.ProposalID

type ProposalState = CallState

// Proposal is a proposal operation
type Proposal[I, O any] Call[ProposalID, I, O]

// Proposals provides access to pending proposals
type Proposals interface {
	// Get gets a proposal by ID
	Get(id ProposalID) (PrimitiveProposal, bool)
	// List lists all open proposals
	List() []PrimitiveProposal
}

func newPrimitiveProposals() *primitiveProposals {
	return &primitiveProposals{
		proposals: make(map[statemachine.ProposalID]*primitiveProposal),
	}
}

type primitiveProposals struct {
	proposals map[statemachine.ProposalID]*primitiveProposal
}

func (p *primitiveProposals) Get(id statemachine.ProposalID) (PrimitiveProposal, bool) {
	proposal, ok := p.proposals[id]
	return proposal, ok
}

func (p *primitiveProposals) List() []PrimitiveProposal {
	proposals := make([]PrimitiveProposal, 0, len(p.proposals))
	for _, proposal := range p.proposals {
		proposals = append(proposals, proposal)
	}
	return proposals
}

func (p *primitiveProposals) add(proposal *primitiveProposal) {
	p.proposals[proposal.ID()] = proposal
}

func (p *primitiveProposals) remove(id statemachine.ProposalID) {
	delete(p.proposals, id)
}

func newSessionProposal(session *managedSession) *sessionProposal {
	return &sessionProposal{
		session: session,
		outputs: list.New(),
	}
}

type sessionProposal struct {
	session      *managedSession
	id           statemachine.ProposalID
	input        *multiraftv1.SessionProposalInput
	timestamp    time.Time
	state        ProposalState
	parent       statemachine.Proposal[*multiraftv1.SessionProposalInput, *multiraftv1.SessionProposalOutput]
	watchers     map[uuid.UUID]func(ProposalState)
	outputs      *list.List
	outputSeqNum multiraftv1.SequenceNum
	log          logging.Logger
}

func (p *sessionProposal) ID() statemachine.ProposalID {
	return p.id
}

func (p *sessionProposal) Log() logging.Logger {
	return p.log
}

func (p *sessionProposal) Session() Session {
	return p.session
}

func (p *sessionProposal) Time() time.Time {
	return p.timestamp
}

func (p *sessionProposal) State() ProposalState {
	return p.state
}

func (p *sessionProposal) Watch(watcher func(ProposalState)) CancelFunc {
	if p.watchers == nil {
		p.watchers = make(map[uuid.UUID]func(ProposalState))
	}
	id := uuid.New()
	p.watchers[id] = watcher
	return func() {
		delete(p.watchers, id)
	}
}

func (p *sessionProposal) execute(parent statemachine.Proposal[*multiraftv1.SessionProposalInput, *multiraftv1.SessionProposalOutput]) {
	p.id = parent.ID()
	p.input = parent.Input()
	p.timestamp = p.session.manager.Time()
	p.state = Running
	p.log = p.session.Log().WithFields(logging.Uint64("ProposalID", uint64(parent.ID())))
	p.parent = parent

	switch parent.Input().Input.(type) {
	case *multiraftv1.SessionProposalInput_Proposal:
		proposal := newPrimitiveProposal(p)
		p.session.manager.proposals.add(proposal)
		p.session.manager.sm.Propose(proposal)
	case *multiraftv1.SessionProposalInput_CreatePrimitive:
		p.session.manager.sm.CreatePrimitive(newCreatePrimitiveProposal(p))
	case *multiraftv1.SessionProposalInput_ClosePrimitive:
		p.session.manager.sm.ClosePrimitive(newClosePrimitiveProposal(p))
	}
}

func (p *sessionProposal) replay(parent statemachine.Proposal[*multiraftv1.SessionProposalInput, *multiraftv1.SessionProposalOutput]) {
	p.parent = parent
	if p.outputs.Len() > 0 {
		p.Log().Debug("Replaying proposal outputs")
		elem := p.outputs.Front()
		for elem != nil {
			output := elem.Value.(*multiraftv1.SessionProposalOutput)
			p.parent.Output(output)
			elem = elem.Next()
		}
	}
	if p.state == Complete {
		p.parent.Close()
	}
}

func (p *sessionProposal) snapshot(writer *snapshot.Writer) error {
	p.Log().Info("Persisting proposal to snapshot")
	pendingOutputs := make([]*multiraftv1.SessionProposalOutput, 0, p.outputs.Len())
	elem := p.outputs.Front()
	for elem != nil {
		pendingOutputs = append(pendingOutputs, elem.Value.(*multiraftv1.SessionProposalOutput))
		elem = elem.Next()
	}

	var phase multiraftv1.SessionProposalSnapshot_Phase
	switch p.state {
	case Pending:
		phase = multiraftv1.SessionProposalSnapshot_PENDING
	case Running:
		phase = multiraftv1.SessionProposalSnapshot_RUNNING
	case Canceled:
		phase = multiraftv1.SessionProposalSnapshot_CANCELED
	case Complete:
		phase = multiraftv1.SessionProposalSnapshot_COMPLETE
	}

	snapshot := &multiraftv1.SessionProposalSnapshot{
		Index:                 multiraftv1.Index(p.ID()),
		Phase:                 phase,
		Input:                 p.input,
		PendingOutputs:        pendingOutputs,
		LastOutputSequenceNum: p.outputSeqNum,
		Timestamp:             p.timestamp,
	}
	return writer.WriteMessage(snapshot)
}

func (p *sessionProposal) recover(reader *snapshot.Reader) error {
	snapshot := &multiraftv1.SessionProposalSnapshot{}
	if err := reader.ReadMessage(snapshot); err != nil {
		return err
	}
	p.id = statemachine.ProposalID(snapshot.Index)
	p.input = snapshot.Input
	p.timestamp = snapshot.Timestamp
	p.log = p.session.Log().WithFields(logging.Uint64("ProposalID", uint64(snapshot.Index)))
	p.Log().Info("Recovering command from snapshot")
	p.outputs = list.New()
	for _, output := range snapshot.PendingOutputs {
		r := output
		p.outputs.PushBack(r)
	}
	p.outputSeqNum = snapshot.LastOutputSequenceNum

	switch snapshot.Phase {
	case multiraftv1.SessionProposalSnapshot_PENDING:
		p.state = Pending
	case multiraftv1.SessionProposalSnapshot_RUNNING:
		p.state = Running
		switch p.input.Input.(type) {
		case *multiraftv1.SessionProposalInput_Proposal:
			proposal := newPrimitiveProposal(p)
			p.session.manager.proposals.add(proposal)
		}
	case multiraftv1.SessionProposalSnapshot_COMPLETE:
		p.state = Complete
	case multiraftv1.SessionProposalSnapshot_CANCELED:
		p.state = Canceled
	}
	return nil
}

func (p *sessionProposal) ack(outputSequenceNum multiraftv1.SequenceNum) {
	p.Log().Debugw("Acked proposal outputs",
		logging.Uint64("SequenceNum", uint64(outputSequenceNum)))
	elem := p.outputs.Front()
	for elem != nil && elem.Value.(*multiraftv1.SessionProposalOutput).SequenceNum <= outputSequenceNum {
		next := elem.Next()
		p.outputs.Remove(elem)
		elem = next
	}
}

func (p *sessionProposal) nextSequenceNum() multiraftv1.SequenceNum {
	p.outputSeqNum++
	return p.outputSeqNum
}

func (p *sessionProposal) Input() *multiraftv1.SessionProposalInput {
	return p.input
}

func (p *sessionProposal) Output(output *multiraftv1.SessionProposalOutput) {
	if p.state != Running {
		return
	}
	p.Log().Debugw("Cached command output", logging.Uint64("SequenceNum", uint64(output.SequenceNum)))
	p.outputs.PushBack(output)
	if p.parent != nil {
		p.parent.Output(output)
	}
}

func (p *sessionProposal) Error(err error) {
	if p.state != Running {
		return
	}
	p.Output(&multiraftv1.SessionProposalOutput{
		SequenceNum: p.nextSequenceNum(),
		Failure:     getFailure(err),
	})
	p.Close()
}

func (p *sessionProposal) Close() {
	p.close(Complete)
}

func (p *sessionProposal) Cancel() {
	p.close(Canceled)
}

func (p *sessionProposal) close(phase ProposalState) {
	if p.state != Running {
		return
	}
	if p.parent != nil {
		p.parent.Close()
	}
	p.state = phase
	p.session.manager.proposals.remove(p.id)
	if p.watchers != nil {
		for _, watcher := range p.watchers {
			watcher(phase)
		}
	}
}

func newPrimitiveProposal(parent *sessionProposal) *primitiveProposal {
	return &primitiveProposal{
		sessionProposal: parent,
	}
}

type primitiveProposal struct {
	*sessionProposal
}

func (p *primitiveProposal) Input() *multiraftv1.PrimitiveProposalInput {
	return p.sessionProposal.Input().GetProposal()
}

func (p *primitiveProposal) Output(output *multiraftv1.PrimitiveProposalOutput) {
	p.sessionProposal.Output(&multiraftv1.SessionProposalOutput{
		SequenceNum: p.nextSequenceNum(),
		Output: &multiraftv1.SessionProposalOutput_Proposal{
			Proposal: output,
		},
	})
}

var _ Proposal[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput] = (*primitiveProposal)(nil)

func newCreatePrimitiveProposal(parent *sessionProposal) *createPrimitiveProposal {
	return &createPrimitiveProposal{
		sessionProposal: parent,
	}
}

type createPrimitiveProposal struct {
	*sessionProposal
}

func (p *createPrimitiveProposal) Input() *multiraftv1.CreatePrimitiveInput {
	return p.sessionProposal.Input().GetCreatePrimitive()
}

func (p *createPrimitiveProposal) Output(output *multiraftv1.CreatePrimitiveOutput) {
	p.sessionProposal.Output(&multiraftv1.SessionProposalOutput{
		SequenceNum: p.nextSequenceNum(),
		Output: &multiraftv1.SessionProposalOutput_CreatePrimitive{
			CreatePrimitive: output,
		},
	})
}

var _ Proposal[*multiraftv1.CreatePrimitiveInput, *multiraftv1.CreatePrimitiveOutput] = (*createPrimitiveProposal)(nil)

func newClosePrimitiveProposal(parent *sessionProposal) *closePrimitiveProposal {
	return &closePrimitiveProposal{
		sessionProposal: parent,
	}
}

type closePrimitiveProposal struct {
	*sessionProposal
}

func (p *closePrimitiveProposal) Input() *multiraftv1.ClosePrimitiveInput {
	return p.sessionProposal.Input().GetClosePrimitive()
}

func (p *closePrimitiveProposal) Output(output *multiraftv1.ClosePrimitiveOutput) {
	p.sessionProposal.Output(&multiraftv1.SessionProposalOutput{
		SequenceNum: p.nextSequenceNum(),
		Output: &multiraftv1.SessionProposalOutput_ClosePrimitive{
			ClosePrimitive: output,
		},
	})
}

var _ Proposal[*multiraftv1.ClosePrimitiveInput, *multiraftv1.ClosePrimitiveOutput] = (*closePrimitiveProposal)(nil)
