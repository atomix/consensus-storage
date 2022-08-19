// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"context"
	api "github.com/atomix/multi-raft-storage/api/atomix/multiraft/countermap/v1"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/driver/pkg/client"
	"github.com/atomix/multi-raft-storage/driver/pkg/util/async"
	atomiccountermapv1 "github.com/atomix/runtime/api/atomix/runtime/countermap/v1"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/atomix/runtime/sdk/pkg/logging"
	"github.com/atomix/runtime/sdk/pkg/runtime"
	"google.golang.org/grpc"
	"io"
)

var log = logging.GetLogger()

const Service = "atomix.multiraft.countermap.v1.CounterMap"

func NewCounterMapServer(protocol *client.Protocol) atomiccountermapv1.CounterMapServer {
	return &multiRaftCounterMapServer{
		Protocol: protocol,
	}
}

type multiRaftCounterMapServer struct {
	*client.Protocol
}

func (s *multiRaftCounterMapServer) Create(ctx context.Context, request *atomiccountermapv1.CreateRequest) (*atomiccountermapv1.CreateResponse, error) {
	log.Debugw("Create",
		logging.Stringer("CreateRequest", request))
	partitions := s.Partitions()
	err := async.IterAsync(len(partitions), func(i int) error {
		partition := partitions[i]
		session, err := partition.GetSession(ctx)
		if err != nil {
			return err
		}
		spec := multiraftv1.PrimitiveSpec{
			Service:   Service,
			Namespace: runtime.GetNamespace(),
			Name:      request.ID.Name,
		}
		return session.CreatePrimitive(ctx, spec)
	})
	if err != nil {
		log.Warnw("Create",
			logging.Stringer("CreateRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &atomiccountermapv1.CreateResponse{}
	log.Debugw("Create",
		logging.Stringer("CreateRequest", request),
		logging.Stringer("CreateResponse", response))
	return response, nil
}

func (s *multiRaftCounterMapServer) Close(ctx context.Context, request *atomiccountermapv1.CloseRequest) (*atomiccountermapv1.CloseResponse, error) {
	log.Debugw("Close",
		logging.Stringer("CloseRequest", request))
	partitions := s.Partitions()
	err := async.IterAsync(len(partitions), func(i int) error {
		partition := partitions[i]
		session, err := partition.GetSession(ctx)
		if err != nil {
			return err
		}
		return session.ClosePrimitive(ctx, request.ID.Name)
	})
	if err != nil {
		log.Warnw("Close",
			logging.Stringer("CloseRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &atomiccountermapv1.CloseResponse{}
	log.Debugw("Close",
		logging.Stringer("CloseRequest", request),
		logging.Stringer("CloseResponse", response))
	return response, nil
}

func (s *multiRaftCounterMapServer) Size(ctx context.Context, request *atomiccountermapv1.SizeRequest) (*atomiccountermapv1.SizeResponse, error) {
	log.Debugw("Size",
		logging.Stringer("SizeRequest", request))
	partitions := s.Partitions()
	sizes, err := async.ExecuteAsync[int](len(partitions), func(i int) (int, error) {
		partition := partitions[i]
		session, err := partition.GetSession(ctx)
		if err != nil {
			log.Warnw("Size",
				logging.Stringer("SizeRequest", request),
				logging.Error("Error", err))
			return 0, err
		}
		primitive, err := session.GetPrimitive(request.ID.Name)
		if err != nil {
			log.Warnw("Size",
				logging.Stringer("SizeRequest", request),
				logging.Error("Error", err))
			return 0, err
		}
		query := client.Query[*api.SizeResponse](primitive)
		output, err := query.Run(func(conn *grpc.ClientConn, headers *multiraftv1.QueryRequestHeaders) (*api.SizeResponse, error) {
			return api.NewCounterMapClient(conn).Size(ctx, &api.SizeRequest{
				Headers:   headers,
				SizeInput: &api.SizeInput{},
			})
		})
		if err != nil {
			log.Warnw("Size",
				logging.Stringer("SizeRequest", request),
				logging.Error("Error", err))
			return 0, err
		}
		return int(output.Size_), nil
	})
	if err != nil {
		return nil, errors.ToProto(err)
	}
	var size int
	for _, s := range sizes {
		size += s
	}
	response := &atomiccountermapv1.SizeResponse{
		Size_: uint32(size),
	}
	log.Debugw("Size",
		logging.Stringer("SizeRequest", request),
		logging.Stringer("SizeResponse", response))
	return response, nil
}

func (s *multiRaftCounterMapServer) Set(ctx context.Context, request *atomiccountermapv1.SetRequest) (*atomiccountermapv1.SetResponse, error) {
	log.Debugw("Set",
		logging.Stringer("SetRequest", request))
	partition := s.PartitionBy([]byte(request.Key))
	session, err := partition.GetSession(ctx)
	if err != nil {
		log.Warnw("Set",
			logging.Stringer("SetRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	primitive, err := session.GetPrimitive(request.ID.Name)
	if err != nil {
		log.Warnw("Set",
			logging.Stringer("SetRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	command := client.Command[*api.SetResponse](primitive)
	output, err := command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (*api.SetResponse, error) {
		input := &api.SetRequest{
			Headers: headers,
			SetInput: &api.SetInput{
				Key:   request.Key,
				Value: request.Value,
			},
		}
		return api.NewCounterMapClient(conn).Set(ctx, input)
	})
	if err != nil {
		log.Warnw("Set",
			logging.Stringer("SetRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &atomiccountermapv1.SetResponse{
		PrevValue: output.PrevValue,
	}
	log.Debugw("Set",
		logging.Stringer("SetRequest", request),
		logging.Stringer("SetResponse", response))
	return response, nil
}

func (s *multiRaftCounterMapServer) Insert(ctx context.Context, request *atomiccountermapv1.InsertRequest) (*atomiccountermapv1.InsertResponse, error) {
	log.Debugw("Insert",
		logging.Stringer("InsertRequest", request))
	partition := s.PartitionBy([]byte(request.Key))
	session, err := partition.GetSession(ctx)
	if err != nil {
		log.Warnw("Insert",
			logging.Stringer("InsertRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	primitive, err := session.GetPrimitive(request.ID.Name)
	if err != nil {
		log.Warnw("Insert",
			logging.Stringer("InsertRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	command := client.Command[*api.InsertResponse](primitive)
	_, err = command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (*api.InsertResponse, error) {
		return api.NewCounterMapClient(conn).Insert(ctx, &api.InsertRequest{
			Headers: headers,
			InsertInput: &api.InsertInput{
				Key:   request.Key,
				Value: request.Value,
			},
		})
	})
	if err != nil {
		log.Warnw("Insert",
			logging.Stringer("InsertRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &atomiccountermapv1.InsertResponse{}
	log.Debugw("Insert",
		logging.Stringer("InsertRequest", request),
		logging.Stringer("InsertResponse", response))
	return response, nil
}

func (s *multiRaftCounterMapServer) Update(ctx context.Context, request *atomiccountermapv1.UpdateRequest) (*atomiccountermapv1.UpdateResponse, error) {
	log.Debugw("Update",
		logging.Stringer("UpdateRequest", request))
	partition := s.PartitionBy([]byte(request.Key))
	session, err := partition.GetSession(ctx)
	if err != nil {
		log.Warnw("Update",
			logging.Stringer("UpdateRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	primitive, err := session.GetPrimitive(request.ID.Name)
	if err != nil {
		log.Warnw("Update",
			logging.Stringer("UpdateRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	command := client.Command[*api.UpdateResponse](primitive)
	output, err := command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (*api.UpdateResponse, error) {
		input := &api.UpdateRequest{
			Headers: headers,
			UpdateInput: &api.UpdateInput{
				Key:   request.Key,
				Value: request.Value,
			},
		}
		return api.NewCounterMapClient(conn).Update(ctx, input)
	})
	if err != nil {
		log.Warnw("Update",
			logging.Stringer("UpdateRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &atomiccountermapv1.UpdateResponse{
		PrevValue: output.PrevValue,
	}
	log.Debugw("Update",
		logging.Stringer("UpdateRequest", request),
		logging.Stringer("UpdateResponse", response))
	return response, nil
}

func (s *multiRaftCounterMapServer) Increment(ctx context.Context, request *atomiccountermapv1.IncrementRequest) (*atomiccountermapv1.IncrementResponse, error) {
	log.Debugw("Increment",
		logging.Stringer("IncrementRequest", request))
	partition := s.PartitionBy([]byte(request.Key))
	session, err := partition.GetSession(ctx)
	if err != nil {
		log.Warnw("Increment",
			logging.Stringer("IncrementRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	primitive, err := session.GetPrimitive(request.ID.Name)
	if err != nil {
		log.Warnw("Increment",
			logging.Stringer("IncrementRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	command := client.Command[*api.IncrementResponse](primitive)
	output, err := command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (*api.IncrementResponse, error) {
		input := &api.IncrementRequest{
			Headers: headers,
			IncrementInput: &api.IncrementInput{
				Key:   request.Key,
				Delta: request.Delta,
			},
		}
		return api.NewCounterMapClient(conn).Increment(ctx, input)
	})
	if err != nil {
		log.Warnw("Increment",
			logging.Stringer("IncrementRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &atomiccountermapv1.IncrementResponse{
		PrevValue: output.PrevValue,
	}
	log.Debugw("Increment",
		logging.Stringer("IncrementRequest", request),
		logging.Stringer("IncrementResponse", response))
	return response, nil
}

func (s *multiRaftCounterMapServer) Decrement(ctx context.Context, request *atomiccountermapv1.DecrementRequest) (*atomiccountermapv1.DecrementResponse, error) {
	log.Debugw("Decrement",
		logging.Stringer("DecrementRequest", request))
	partition := s.PartitionBy([]byte(request.Key))
	session, err := partition.GetSession(ctx)
	if err != nil {
		log.Warnw("Decrement",
			logging.Stringer("DecrementRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	primitive, err := session.GetPrimitive(request.ID.Name)
	if err != nil {
		log.Warnw("Decrement",
			logging.Stringer("DecrementRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	command := client.Command[*api.DecrementResponse](primitive)
	output, err := command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (*api.DecrementResponse, error) {
		input := &api.DecrementRequest{
			Headers: headers,
			DecrementInput: &api.DecrementInput{
				Key:   request.Key,
				Delta: request.Delta,
			},
		}
		return api.NewCounterMapClient(conn).Decrement(ctx, input)
	})
	if err != nil {
		log.Warnw("Decrement",
			logging.Stringer("DecrementRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &atomiccountermapv1.DecrementResponse{
		PrevValue: output.PrevValue,
	}
	log.Debugw("Decrement",
		logging.Stringer("DecrementRequest", request),
		logging.Stringer("DecrementResponse", response))
	return response, nil
}

func (s *multiRaftCounterMapServer) Get(ctx context.Context, request *atomiccountermapv1.GetRequest) (*atomiccountermapv1.GetResponse, error) {
	log.Debugw("Get",
		logging.Stringer("GetRequest", request))
	partition := s.PartitionBy([]byte(request.Key))
	session, err := partition.GetSession(ctx)
	if err != nil {
		log.Warnw("Get",
			logging.Stringer("GetRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	primitive, err := session.GetPrimitive(request.ID.Name)
	if err != nil {
		log.Warnw("Get",
			logging.Stringer("GetRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	query := client.Query[*api.GetResponse](primitive)
	output, err := query.Run(func(conn *grpc.ClientConn, headers *multiraftv1.QueryRequestHeaders) (*api.GetResponse, error) {
		return api.NewCounterMapClient(conn).Get(ctx, &api.GetRequest{
			Headers: headers,
			GetInput: &api.GetInput{
				Key: request.Key,
			},
		})
	})
	if err != nil {
		log.Warnw("Get",
			logging.Stringer("GetRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &atomiccountermapv1.GetResponse{
		Value: output.Value,
	}
	log.Debugw("Get",
		logging.Stringer("GetRequest", request),
		logging.Stringer("GetResponse", response))
	return response, nil
}

func (s *multiRaftCounterMapServer) Remove(ctx context.Context, request *atomiccountermapv1.RemoveRequest) (*atomiccountermapv1.RemoveResponse, error) {
	log.Debugw("Remove",
		logging.Stringer("RemoveRequest", request))
	partition := s.PartitionBy([]byte(request.Key))
	session, err := partition.GetSession(ctx)
	if err != nil {
		log.Warnw("Remove",
			logging.Stringer("RemoveRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	primitive, err := session.GetPrimitive(request.ID.Name)
	if err != nil {
		log.Warnw("Remove",
			logging.Stringer("RemoveRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	command := client.Command[*api.RemoveResponse](primitive)
	output, err := command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (*api.RemoveResponse, error) {
		input := &api.RemoveRequest{
			Headers: headers,
			RemoveInput: &api.RemoveInput{
				Key:       request.Key,
				PrevValue: request.PrevValue,
			},
		}
		return api.NewCounterMapClient(conn).Remove(ctx, input)
	})
	if err != nil {
		log.Warnw("Remove",
			logging.Stringer("RemoveRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &atomiccountermapv1.RemoveResponse{
		Value: output.Value,
	}
	log.Debugw("Remove",
		logging.Stringer("RemoveRequest", request),
		logging.Stringer("RemoveResponse", response))
	return response, nil
}

func (s *multiRaftCounterMapServer) Clear(ctx context.Context, request *atomiccountermapv1.ClearRequest) (*atomiccountermapv1.ClearResponse, error) {
	log.Debugw("Clear",
		logging.Stringer("ClearRequest", request))
	partitions := s.Partitions()
	err := async.IterAsync(len(partitions), func(i int) error {
		partition := partitions[i]
		session, err := partition.GetSession(ctx)
		if err != nil {
			log.Warnw("Clear",
				logging.Stringer("ClearRequest", request),
				logging.Error("Error", err))
			return err
		}
		primitive, err := session.GetPrimitive(request.ID.Name)
		if err != nil {
			log.Warnw("Clear",
				logging.Stringer("ClearRequest", request),
				logging.Error("Error", err))
			return err
		}
		command := client.Command[*api.ClearResponse](primitive)
		_, err = command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (*api.ClearResponse, error) {
			return api.NewCounterMapClient(conn).Clear(ctx, &api.ClearRequest{
				Headers:    headers,
				ClearInput: &api.ClearInput{},
			})
		})
		if err != nil {
			log.Warnw("Clear",
				logging.Stringer("ClearRequest", request),
				logging.Error("Error", err))
			return err
		}
		return nil
	})
	if err != nil {
		return nil, errors.ToProto(err)
	}
	response := &atomiccountermapv1.ClearResponse{}
	log.Debugw("Clear",
		logging.Stringer("ClearRequest", request),
		logging.Stringer("ClearResponse", response))
	return response, nil
}

func (s *multiRaftCounterMapServer) Lock(ctx context.Context, request *atomiccountermapv1.LockRequest) (*atomiccountermapv1.LockResponse, error) {
	log.Debugw("Lock",
		logging.Stringer("LockRequest", request))

	partitions := s.Partitions()
	indexKeys := make(map[int][]string)
	for _, key := range request.Keys {
		index := s.PartitionIndex([]byte(key))
		indexKeys[index] = append(indexKeys[index], key)
	}

	partitionIndexes := make([]int, 0, len(indexKeys))
	for index := range indexKeys {
		partitionIndexes = append(partitionIndexes, index)
	}

	err := async.IterAsync(len(partitionIndexes), func(i int) error {
		index := partitionIndexes[i]
		partition := partitions[index]
		keys := indexKeys[index]

		session, err := partition.GetSession(ctx)
		if err != nil {
			log.Warnw("Lock",
				logging.Stringer("LockRequest", request),
				logging.Error("Error", err))
			return err
		}
		primitive, err := session.GetPrimitive(request.ID.Name)
		if err != nil {
			log.Warnw("Lock",
				logging.Stringer("LockRequest", request),
				logging.Error("Error", err))
			return err
		}
		command := client.Command[*api.LockResponse](primitive)
		_, err = command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (*api.LockResponse, error) {
			return api.NewCounterMapClient(conn).Lock(ctx, &api.LockRequest{
				Headers: headers,
				LockInput: &api.LockInput{
					Keys:    keys,
					Timeout: request.Timeout,
				},
			})
		})
		if err != nil {
			log.Warnw("Lock",
				logging.Stringer("LockRequest", request),
				logging.Error("Error", err))
			return err
		}
		return nil
	})
	if err != nil {
		return nil, errors.ToProto(err)
	}
	response := &atomiccountermapv1.LockResponse{}
	log.Debugw("Lock",
		logging.Stringer("LockRequest", request),
		logging.Stringer("LockResponse", response))
	return response, nil
}

func (s *multiRaftCounterMapServer) Unlock(ctx context.Context, request *atomiccountermapv1.UnlockRequest) (*atomiccountermapv1.UnlockResponse, error) {
	log.Debugw("Unlock",
		logging.Stringer("UnlockRequest", request))
	partitions := s.Partitions()
	err := async.IterAsync(len(partitions), func(i int) error {
		partition := partitions[i]
		session, err := partition.GetSession(ctx)
		if err != nil {
			log.Warnw("Unlock",
				logging.Stringer("UnlockRequest", request),
				logging.Error("Error", err))
			return err
		}
		primitive, err := session.GetPrimitive(request.ID.Name)
		if err != nil {
			log.Warnw("Unlock",
				logging.Stringer("UnlockRequest", request),
				logging.Error("Error", err))
			return err
		}
		command := client.Command[*api.UnlockResponse](primitive)
		_, err = command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (*api.UnlockResponse, error) {
			return api.NewCounterMapClient(conn).Unlock(ctx, &api.UnlockRequest{
				Headers:     headers,
				UnlockInput: &api.UnlockInput{},
			})
		})
		if err != nil {
			log.Warnw("Unlock",
				logging.Stringer("UnlockRequest", request),
				logging.Error("Error", err))
			return err
		}
		return nil
	})
	if err != nil {
		return nil, errors.ToProto(err)
	}
	response := &atomiccountermapv1.UnlockResponse{}
	log.Debugw("Unlock",
		logging.Stringer("UnlockRequest", request),
		logging.Stringer("UnlockResponse", response))
	return response, nil
}

func (s *multiRaftCounterMapServer) Events(request *atomiccountermapv1.EventsRequest, server atomiccountermapv1.CounterMap_EventsServer) error {
	log.Debugw("Events",
		logging.Stringer("EventsRequest", request))
	partitions := s.Partitions()
	return async.IterAsync(len(partitions), func(i int) error {
		partition := partitions[i]
		session, err := partition.GetSession(server.Context())
		if err != nil {
			err = errors.ToProto(err)
			log.Warnw("Events",
				logging.Stringer("EventsRequest", request),
				logging.Error("Error", err))
			return err
		}
		primitive, err := session.GetPrimitive(request.ID.Name)
		if err != nil {
			err = errors.ToProto(err)
			log.Warnw("Events",
				logging.Stringer("EventsRequest", request),
				logging.Error("Error", err))
			return err
		}
		command := client.StreamCommand[api.CounterMap_EventsClient, *api.EventsResponse](primitive)
		stream, err := command.Open(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (api.CounterMap_EventsClient, error) {
			return api.NewCounterMapClient(conn).Events(server.Context(), &api.EventsRequest{
				Headers: headers,
				EventsInput: &api.EventsInput{
					Key: request.Key,
				},
			})
		})
		if err != nil {
			err = errors.ToProto(err)
			log.Warnw("Events",
				logging.Stringer("EventsRequest", request),
				logging.Error("Error", err))
			return err
		}
		for {
			output, err := command.Recv(stream.Recv)
			if err == io.EOF {
				log.Debugw("Events",
					logging.Stringer("EventsRequest", request),
					logging.String("State", "Done"))
				return nil
			}
			if err != nil {
				log.Warnw("Events",
					logging.Stringer("EventsRequest", request),
					logging.Error("Error", err))
				return errors.ToProto(err)
			}
			var response *atomiccountermapv1.EventsResponse
			switch e := output.Event.Event.(type) {
			case *api.Event_Inserted_:
				response = &atomiccountermapv1.EventsResponse{
					Event: atomiccountermapv1.Event{
						Key: output.Event.Key,
						Event: &atomiccountermapv1.Event_Inserted_{
							Inserted: &atomiccountermapv1.Event_Inserted{
								Value: e.Inserted.Value,
							},
						},
					},
				}
			case *api.Event_Updated_:
				response = &atomiccountermapv1.EventsResponse{
					Event: atomiccountermapv1.Event{
						Key: output.Event.Key,
						Event: &atomiccountermapv1.Event_Updated_{
							Updated: &atomiccountermapv1.Event_Updated{
								Value:     e.Updated.Value,
								PrevValue: e.Updated.PrevValue,
							},
						},
					},
				}
			case *api.Event_Removed_:
				response = &atomiccountermapv1.EventsResponse{
					Event: atomiccountermapv1.Event{
						Key: output.Event.Key,
						Event: &atomiccountermapv1.Event_Removed_{
							Removed: &atomiccountermapv1.Event_Removed{
								Value: e.Removed.Value,
							},
						},
					},
				}
			}
			log.Debugw("Events",
				logging.Stringer("EventsRequest", request),
				logging.Stringer("EventsResponse", response))
			if err := server.Send(response); err != nil {
				log.Warnw("Events",
					logging.Stringer("EventsRequest", request),
					logging.Stringer("EventsResponse", response),
					logging.Error("Error", err))
				return err
			}
		}
	})
}

func (s *multiRaftCounterMapServer) Entries(request *atomiccountermapv1.EntriesRequest, server atomiccountermapv1.CounterMap_EntriesServer) error {
	log.Debugw("Entries",
		logging.Stringer("EntriesRequest", request))
	partitions := s.Partitions()
	return async.IterAsync(len(partitions), func(i int) error {
		partition := partitions[i]
		session, err := partition.GetSession(server.Context())
		if err != nil {
			log.Warnw("Entries",
				logging.Stringer("EntriesRequest", request),
				logging.Error("Error", err))
			return errors.ToProto(err)
		}
		primitive, err := session.GetPrimitive(request.ID.Name)
		if err != nil {
			log.Warnw("Entries",
				logging.Stringer("EntriesRequest", request),
				logging.Error("Error", err))
			return errors.ToProto(err)
		}
		query := client.StreamQuery[api.CounterMap_EntriesClient, *api.EntriesResponse](primitive)
		stream, err := query.Open(func(conn *grpc.ClientConn, headers *multiraftv1.QueryRequestHeaders) (api.CounterMap_EntriesClient, error) {
			return api.NewCounterMapClient(conn).Entries(server.Context(), &api.EntriesRequest{
				Headers: headers,
				EntriesInput: &api.EntriesInput{
					Watch: request.Watch,
				},
			})
		})
		if err != nil {
			log.Warnw("Entries",
				logging.Stringer("EntriesRequest", request),
				logging.Error("Error", err))
			return errors.ToProto(err)
		}
		for {
			output, err := query.Recv(stream.Recv)
			if err == io.EOF {
				return nil
			}
			if err != nil {
				log.Warnw("Entries",
					logging.Stringer("EntriesRequest", request),
					logging.Error("Error", err))
				return errors.ToProto(err)
			}
			response := &atomiccountermapv1.EntriesResponse{
				Entry: atomiccountermapv1.Entry{
					Key:   output.Entry.Key,
					Value: output.Entry.Value,
				},
			}
			log.Debugw("Entries",
				logging.Stringer("EntriesRequest", request),
				logging.Stringer("EntriesResponse", response))
			if err := server.Send(response); err != nil {
				log.Warnw("Entries",
					logging.Stringer("EntriesRequest", request),
					logging.Stringer("EntriesResponse", response),
					logging.Error("Error", err))
				return err
			}
		}
	})
}

var _ atomiccountermapv1.CounterMapServer = (*multiRaftCounterMapServer)(nil)
