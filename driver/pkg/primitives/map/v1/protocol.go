// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"context"
	api "github.com/atomix/multi-raft-storage/api/atomix/multiraft/map/v1"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/driver/pkg/client"
	"github.com/atomix/multi-raft-storage/driver/pkg/util/async"
	atomicmapv1 "github.com/atomix/runtime/api/atomix/runtime/map/v1"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/atomix/runtime/sdk/pkg/logging"
	"google.golang.org/grpc"
	"io"
)

const Service = "atomix.runtime.map.v1.Map"

func newMultiRaftMapServer(protocol *client.Protocol) atomicmapv1.MapServer {
	return &multiRaftMapServer{
		Protocol: protocol,
	}
}

type multiRaftMapServer struct {
	*client.Protocol
}

func (s *multiRaftMapServer) Create(ctx context.Context, request *atomicmapv1.CreateRequest) (*atomicmapv1.CreateResponse, error) {
	log.Debugw("Create",
		logging.Stringer("CreateRequest", request))
	partitions := s.Partitions()
	err := async.IterAsync(len(partitions), func(i int) error {
		partition := partitions[i]
		session, err := partition.GetSession(ctx)
		if err != nil {
			return err
		}
		return session.CreatePrimitive(ctx, request.ID.Name, Service)
	})
	if err != nil {
		log.Warnw("Create",
			logging.Stringer("CreateRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &atomicmapv1.CreateResponse{}
	log.Debugw("Create",
		logging.Stringer("CreateRequest", request),
		logging.Stringer("CreateResponse", response))
	return response, nil
}

func (s *multiRaftMapServer) Close(ctx context.Context, request *atomicmapv1.CloseRequest) (*atomicmapv1.CloseResponse, error) {
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
	response := &atomicmapv1.CloseResponse{}
	log.Debugw("Close",
		logging.Stringer("CloseRequest", request),
		logging.Stringer("CloseResponse", response))
	return response, nil
}

func (s *multiRaftMapServer) Size(ctx context.Context, request *atomicmapv1.SizeRequest) (*atomicmapv1.SizeResponse, error) {
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
			return api.NewMapClient(conn).Size(ctx, &api.SizeRequest{
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
	response := &atomicmapv1.SizeResponse{
		Size_: uint32(size),
	}
	log.Debugw("Size",
		logging.Stringer("SizeRequest", request),
		logging.Stringer("SizeResponse", response))
	return response, nil
}

func (s *multiRaftMapServer) Put(ctx context.Context, request *atomicmapv1.PutRequest) (*atomicmapv1.PutResponse, error) {
	log.Debugw("Put",
		logging.Stringer("PutRequest", request))
	partition := s.PartitionBy([]byte(request.Key))
	session, err := partition.GetSession(ctx)
	if err != nil {
		log.Warnw("Put",
			logging.Stringer("PutRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	primitive, err := session.GetPrimitive(request.ID.Name)
	if err != nil {
		log.Warnw("Put",
			logging.Stringer("PutRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	command := client.Command[*api.PutResponse](primitive)
	output, err := command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (*api.PutResponse, error) {
		input := &api.PutRequest{
			Headers: headers,
			PutInput: &api.PutInput{
				Key:       request.Key,
				Value:     request.Value,
				TTL:       request.TTL,
				PrevIndex: multiraftv1.Index(request.PrevVersion),
			},
		}
		return api.NewMapClient(conn).Put(ctx, input)
	})
	if err != nil {
		log.Warnw("Put",
			logging.Stringer("PutRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &atomicmapv1.PutResponse{
		Version: uint64(output.Index),
	}
	if output.PrevValue != nil {
		response.PrevValue = &atomicmapv1.VersionedValue{
			Value:   output.PrevValue.Value,
			Version: uint64(output.PrevValue.Index),
		}
	}
	log.Debugw("Put",
		logging.Stringer("PutRequest", request),
		logging.Stringer("PutResponse", response))
	return response, nil
}

func (s *multiRaftMapServer) Insert(ctx context.Context, request *atomicmapv1.InsertRequest) (*atomicmapv1.InsertResponse, error) {
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
	output, err := command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (*api.InsertResponse, error) {
		return api.NewMapClient(conn).Insert(ctx, &api.InsertRequest{
			Headers: headers,
			InsertInput: &api.InsertInput{
				Key:   request.Key,
				Value: request.Value,
				TTL:   request.TTL,
			},
		})
	})
	if err != nil {
		log.Warnw("Insert",
			logging.Stringer("InsertRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &atomicmapv1.InsertResponse{
		Version: uint64(output.Index),
	}
	log.Debugw("Insert",
		logging.Stringer("InsertRequest", request),
		logging.Stringer("InsertResponse", response))
	return response, nil
}

func (s *multiRaftMapServer) Update(ctx context.Context, request *atomicmapv1.UpdateRequest) (*atomicmapv1.UpdateResponse, error) {
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
				Key:       request.Key,
				Value:     request.Value,
				TTL:       request.TTL,
				PrevIndex: multiraftv1.Index(request.PrevVersion),
			},
		}
		return api.NewMapClient(conn).Update(ctx, input)
	})
	if err != nil {
		log.Warnw("Update",
			logging.Stringer("UpdateRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &atomicmapv1.UpdateResponse{
		Version: uint64(output.Index),
		PrevValue: atomicmapv1.VersionedValue{
			Value:   output.PrevValue.Value,
			Version: uint64(output.PrevValue.Index),
		},
	}
	log.Debugw("Update",
		logging.Stringer("UpdateRequest", request),
		logging.Stringer("UpdateResponse", response))
	return response, nil
}

func (s *multiRaftMapServer) Get(ctx context.Context, request *atomicmapv1.GetRequest) (*atomicmapv1.GetResponse, error) {
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
		return api.NewMapClient(conn).Get(ctx, &api.GetRequest{
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
	response := &atomicmapv1.GetResponse{
		Value: atomicmapv1.VersionedValue{
			Value:   output.Value.Value,
			Version: uint64(output.Value.Index),
		},
	}
	log.Debugw("Get",
		logging.Stringer("GetRequest", request),
		logging.Stringer("GetResponse", response))
	return response, nil
}

func (s *multiRaftMapServer) Remove(ctx context.Context, request *atomicmapv1.RemoveRequest) (*atomicmapv1.RemoveResponse, error) {
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
				PrevIndex: multiraftv1.Index(request.PrevVersion),
			},
		}
		return api.NewMapClient(conn).Remove(ctx, input)
	})
	if err != nil {
		log.Warnw("Remove",
			logging.Stringer("RemoveRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &atomicmapv1.RemoveResponse{
		Value: atomicmapv1.VersionedValue{
			Value:   output.Value.Value,
			Version: uint64(output.Value.Index),
		},
	}
	log.Debugw("Remove",
		logging.Stringer("RemoveRequest", request),
		logging.Stringer("RemoveResponse", response))
	return response, nil
}

func (s *multiRaftMapServer) Clear(ctx context.Context, request *atomicmapv1.ClearRequest) (*atomicmapv1.ClearResponse, error) {
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
			return api.NewMapClient(conn).Clear(ctx, &api.ClearRequest{
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
	response := &atomicmapv1.ClearResponse{}
	log.Debugw("Clear",
		logging.Stringer("ClearRequest", request),
		logging.Stringer("ClearResponse", response))
	return response, nil
}

func (s *multiRaftMapServer) Lock(ctx context.Context, request *atomicmapv1.LockRequest) (*atomicmapv1.LockResponse, error) {
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
			return api.NewMapClient(conn).Lock(ctx, &api.LockRequest{
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
	response := &atomicmapv1.LockResponse{}
	log.Debugw("Lock",
		logging.Stringer("LockRequest", request),
		logging.Stringer("LockResponse", response))
	return response, nil
}

func (s *multiRaftMapServer) Unlock(ctx context.Context, request *atomicmapv1.UnlockRequest) (*atomicmapv1.UnlockResponse, error) {
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
			return api.NewMapClient(conn).Unlock(ctx, &api.UnlockRequest{
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
	response := &atomicmapv1.UnlockResponse{}
	log.Debugw("Unlock",
		logging.Stringer("UnlockRequest", request),
		logging.Stringer("UnlockResponse", response))
	return response, nil
}

func (s *multiRaftMapServer) Events(request *atomicmapv1.EventsRequest, server atomicmapv1.Map_EventsServer) error {
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
		command := client.StreamCommand[*api.EventsResponse](primitive)
		stream, err := command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (client.CommandStream[*api.EventsResponse], error) {
			return api.NewMapClient(conn).Events(server.Context(), &api.EventsRequest{
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
			output, err := stream.Recv()
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
			var response *atomicmapv1.EventsResponse
			switch e := output.Event.Event.(type) {
			case *api.Event_Inserted_:
				response = &atomicmapv1.EventsResponse{
					Event: atomicmapv1.Event{
						Key: output.Event.Key,
						Event: &atomicmapv1.Event_Inserted_{
							Inserted: &atomicmapv1.Event_Inserted{
								Value: atomicmapv1.VersionedValue{
									Value:   e.Inserted.Value.Value,
									Version: uint64(e.Inserted.Value.Index),
								},
							},
						},
					},
				}
			case *api.Event_Updated_:
				response = &atomicmapv1.EventsResponse{
					Event: atomicmapv1.Event{
						Key: output.Event.Key,
						Event: &atomicmapv1.Event_Updated_{
							Updated: &atomicmapv1.Event_Updated{
								Value: atomicmapv1.VersionedValue{
									Value:   e.Updated.Value.Value,
									Version: uint64(e.Updated.Value.Index),
								},
								PrevValue: atomicmapv1.VersionedValue{
									Value:   e.Updated.PrevValue.Value,
									Version: uint64(e.Updated.PrevValue.Index),
								},
							},
						},
					},
				}
			case *api.Event_Removed_:
				response = &atomicmapv1.EventsResponse{
					Event: atomicmapv1.Event{
						Key: output.Event.Key,
						Event: &atomicmapv1.Event_Removed_{
							Removed: &atomicmapv1.Event_Removed{
								Value: atomicmapv1.VersionedValue{
									Value:   e.Removed.Value.Value,
									Version: uint64(e.Removed.Value.Index),
								},
								Expired: e.Removed.Expired,
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

func (s *multiRaftMapServer) Entries(request *atomicmapv1.EntriesRequest, server atomicmapv1.Map_EntriesServer) error {
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
		query := client.StreamQuery[*api.EntriesResponse](primitive)
		stream, err := query.Run(func(conn *grpc.ClientConn, headers *multiraftv1.QueryRequestHeaders) (client.QueryStream[*api.EntriesResponse], error) {
			return api.NewMapClient(conn).Entries(server.Context(), &api.EntriesRequest{
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
			output, err := stream.Recv()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				log.Warnw("Entries",
					logging.Stringer("EntriesRequest", request),
					logging.Error("Error", err))
				return errors.ToProto(err)
			}
			response := &atomicmapv1.EntriesResponse{
				Entry: atomicmapv1.Entry{
					Key: output.Entry.Key,
				},
			}
			if output.Entry.Value != nil {
				response.Entry.Value = &atomicmapv1.VersionedValue{
					Value:   output.Entry.Value.Value,
					Version: uint64(output.Entry.Value.Index),
				}
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

var _ atomicmapv1.MapServer = (*multiRaftMapServer)(nil)
