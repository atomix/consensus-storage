// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"github.com/atomix/multi-raft-storage/node/pkg/primitive"
	"github.com/atomix/multi-raft-storage/node/pkg/snapshot"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/gogo/protobuf/proto"
)

const (
	setOp           = "Set"
	compareAndSetOp = "CompareAndSet"
	getOp           = "Get"
	incrementOp     = "Increment"
	decrementOp     = "Decrement"
)

type primitiveAdapter struct {
	primitive.Context
	rsm StateMachine
}

func (s *primitiveAdapter) Command(command primitive.Command) {
	switch command.OperationID() {
	case setOp:
		p, err := newSetProposal(command)
		if err != nil {
			err = errors.NewInternal(err.Error())
			log.Error(err)
			command.Output(nil, err)
			return
		}

		log.Debugf("Proposal SetProposal %.250s", p)
		response, err := s.rsm.Set(p)
		if err != nil {
			log.Debugf("Proposal SetProposal %.250s failed: %v", p, err)
			command.Output(nil, err)
		} else {
			output, err := proto.Marshal(response)
			if err != nil {
				err = errors.NewInternal(err.Error())
				log.Errorf("Proposal SetProposal %.250s failed: %v", p, err)
				command.Output(nil, err)
			} else {
				log.Debugf("Proposal SetProposal %.250s complete: %.250s", p, response)
				command.Output(output, nil)
			}
		}
		command.Close()
	case compareAndSetOp:
		p, err := newCompareAndSetProposal(command)
		if err != nil {
			err = errors.NewInternal(err.Error())
			log.Error(err)
			command.Output(nil, err)
			return
		}

		log.Debugf("Proposal CompareAndSetProposal %.250s", p)
		response, err := s.rsm.CompareAndSet(p)
		if err != nil {
			log.Debugf("Proposal CompareAndSetProposal %.250s failed: %v", p, err)
			command.Output(nil, err)
		} else {
			output, err := proto.Marshal(response)
			if err != nil {
				err = errors.NewInternal(err.Error())
				log.Errorf("Proposal CompareAndSetProposal %.250s failed: %v", p, err)
				command.Output(nil, err)
			} else {
				log.Debugf("Proposal CompareAndSetProposal %.250s complete: %.250s", p, response)
				command.Output(output, nil)
			}
		}
		command.Close()
	case incrementOp:
		p, err := newIncrementProposal(command)
		if err != nil {
			err = errors.NewInternal(err.Error())
			log.Error(err)
			command.Output(nil, err)
			return
		}

		log.Debugf("Proposal IncrementProposal %.250s", p)
		response, err := s.rsm.Increment(p)
		if err != nil {
			log.Debugf("Proposal IncrementProposal %.250s failed: %v", p, err)
			command.Output(nil, err)
		} else {
			output, err := proto.Marshal(response)
			if err != nil {
				err = errors.NewInternal(err.Error())
				log.Errorf("Proposal IncrementProposal %.250s failed: %v", p, err)
				command.Output(nil, err)
			} else {
				log.Debugf("Proposal IncrementProposal %.250s complete: %.250s", p, response)
				command.Output(output, nil)
			}
		}
		command.Close()
	case decrementOp:
		p, err := newDecrementProposal(command)
		if err != nil {
			err = errors.NewInternal(err.Error())
			log.Error(err)
			command.Output(nil, err)
			return
		}

		log.Debugf("Proposal DecrementProposal %.250s", p)
		response, err := s.rsm.Decrement(p)
		if err != nil {
			log.Debugf("Proposal DecrementProposal %.250s failed: %v", p, err)
			command.Output(nil, err)
		} else {
			output, err := proto.Marshal(response)
			if err != nil {
				err = errors.NewInternal(err.Error())
				log.Errorf("Proposal DecrementProposal %.250s failed: %v", p, err)
				command.Output(nil, err)
			} else {
				log.Debugf("Proposal DecrementProposal %.250s complete: %.250s", p, response)
				command.Output(output, nil)
			}
		}
		command.Close()
	default:
		err := errors.NewNotSupported("unknown operation %d", command.OperationID())
		log.Debug(err)
		command.Output(nil, err)
	}
}

func (s *primitiveAdapter) Query(query primitive.Query) {
	switch query.OperationID() {
	case getOp:
		q, err := newGetQuery(query)
		if err != nil {
			err = errors.NewInternal(err.Error())
			log.Error(err)
			query.Output(nil, err)
			return
		}

		log.Debugf("Querying GetQuery %.250s", q)
		response, err := s.rsm.Get(q)
		if err != nil {
			log.Debugf("Querying GetQuery %.250s failed: %v", q, err)
			query.Output(nil, err)
		} else {
			output, err := proto.Marshal(response)
			if err != nil {
				err = errors.NewInternal(err.Error())
				log.Errorf("Querying GetQuery %.250s failed: %v", q, err)
				query.Output(nil, err)
			} else {
				log.Debugf("Querying GetQuery %.250s complete: %+v", q, response)
				query.Output(output, nil)
			}
		}
		query.Close()
	default:
		err := errors.NewNotSupported("unknown operation %d", query.OperationID())
		log.Debug(err)
		query.Output(nil, err)
	}
}
func (s *primitiveAdapter) Snapshot(writer *snapshot.Writer) error {
	err := s.rsm.Backup(writer)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (s *primitiveAdapter) Recover(reader *snapshot.Reader) error {
	err := s.rsm.Restore(reader)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

var _ primitive.StateMachine = &primitiveAdapter{}
