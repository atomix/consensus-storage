// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package statemachine

import (
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine2/snapshot"
)

func NewStateMachine(context Context, factory NewSessionManagerFunc) *StateMachine {
	return &StateMachine{
		Context: context,
		sm:      factory(context),
	}
}

type StateMachine struct {
	Context
	sm SessionManager
}

func (s *StateMachine) Snapshot(writer *snapshot.Writer) error {
	return s.sm.Snapshot(writer)
}

func (s *StateMachine) Recover(reader *snapshot.Reader) error {
	return s.sm.Recover(reader)
}

func (s *StateMachine) Propose(proposal Proposal[*multiraftv1.StateMachineProposalInput, *multiraftv1.StateMachineProposalOutput]) {
	switch p := proposal.Input().Input.(type) {
	case *multiraftv1.StateMachineProposalInput_Proposal:
		s.sm.Propose(NewTranscodingProposal[*multiraftv1.StateMachineProposalInput, *multiraftv1.StateMachineProposalOutput, *multiraftv1.SessionProposalInput, *multiraftv1.SessionProposalOutput](
			proposal,
			p.Proposal,
			func(output *multiraftv1.SessionProposalOutput) *multiraftv1.StateMachineProposalOutput {
				return &multiraftv1.StateMachineProposalOutput{
					Index: multiraftv1.Index(s.Index()),
					Output: &multiraftv1.StateMachineProposalOutput_Proposal{
						Proposal: output,
					},
				}
			}))
	case *multiraftv1.StateMachineProposalInput_OpenSession:
		s.sm.OpenSession(NewTranscodingProposal[*multiraftv1.StateMachineProposalInput, *multiraftv1.StateMachineProposalOutput, *multiraftv1.OpenSessionInput, *multiraftv1.OpenSessionOutput](
			proposal,
			p.OpenSession,
			func(output *multiraftv1.OpenSessionOutput) *multiraftv1.StateMachineProposalOutput {
				return &multiraftv1.StateMachineProposalOutput{
					Index: multiraftv1.Index(s.Index()),
					Output: &multiraftv1.StateMachineProposalOutput_OpenSession{
						OpenSession: output,
					},
				}
			}))
	case *multiraftv1.StateMachineProposalInput_KeepAlive:
		s.sm.KeepAlive(NewTranscodingProposal[*multiraftv1.StateMachineProposalInput, *multiraftv1.StateMachineProposalOutput, *multiraftv1.KeepAliveInput, *multiraftv1.KeepAliveOutput](
			proposal,
			p.KeepAlive,
			func(output *multiraftv1.KeepAliveOutput) *multiraftv1.StateMachineProposalOutput {
				return &multiraftv1.StateMachineProposalOutput{
					Index: multiraftv1.Index(s.Index()),
					Output: &multiraftv1.StateMachineProposalOutput_KeepAlive{
						KeepAlive: output,
					},
				}
			}))
	case *multiraftv1.StateMachineProposalInput_CloseSession:
		s.sm.CloseSession(NewTranscodingProposal[*multiraftv1.StateMachineProposalInput, *multiraftv1.StateMachineProposalOutput, *multiraftv1.CloseSessionInput, *multiraftv1.CloseSessionOutput](
			proposal,
			p.CloseSession,
			func(output *multiraftv1.CloseSessionOutput) *multiraftv1.StateMachineProposalOutput {
				return &multiraftv1.StateMachineProposalOutput{
					Index: multiraftv1.Index(s.Index()),
					Output: &multiraftv1.StateMachineProposalOutput_CloseSession{
						CloseSession: output,
					},
				}
			}))
	}
}

func (s *StateMachine) Query(query Query[*multiraftv1.StateMachineQueryInput, *multiraftv1.StateMachineQueryOutput]) {
	minIndex := Index(query.Input().MaxReceivedIndex)
	if s.Index() < minIndex {
		s.Scheduler().Await(minIndex, func() {
			switch q := query.Input().Input.(type) {
			case *multiraftv1.StateMachineQueryInput_Query:
				s.sm.Query(NewTranscodingQuery[*multiraftv1.StateMachineQueryInput, *multiraftv1.StateMachineQueryOutput, *multiraftv1.SessionQueryInput, *multiraftv1.SessionQueryOutput](
					query,
					q.Query,
					func(output *multiraftv1.SessionQueryOutput) *multiraftv1.StateMachineQueryOutput {
						return &multiraftv1.StateMachineQueryOutput{
							Index: multiraftv1.Index(s.Index()),
							Output: &multiraftv1.StateMachineQueryOutput_Query{
								Query: output,
							},
						}
					}))
			}
		})
	} else {
		switch q := query.Input().Input.(type) {
		case *multiraftv1.StateMachineQueryInput_Query:
			s.sm.Query(NewTranscodingQuery[*multiraftv1.StateMachineQueryInput, *multiraftv1.StateMachineQueryOutput, *multiraftv1.SessionQueryInput, *multiraftv1.SessionQueryOutput](
				query,
				q.Query,
				func(output *multiraftv1.SessionQueryOutput) *multiraftv1.StateMachineQueryOutput {
					return &multiraftv1.StateMachineQueryOutput{
						Index: multiraftv1.Index(s.Index()),
						Output: &multiraftv1.StateMachineQueryOutput_Query{
							Query: output,
						},
					}
				}))
		}
	}
}
