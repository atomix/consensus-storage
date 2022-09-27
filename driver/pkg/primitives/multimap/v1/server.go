// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"context"
	api "github.com/atomix/multi-raft-storage/api/atomix/multiraft/multimap/v1"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/driver/pkg/client"
	"github.com/atomix/multi-raft-storage/driver/pkg/util/async"
	atomicmultimapv1 "github.com/atomix/runtime/api/atomix/runtime/multimap/v1"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/atomix/runtime/sdk/pkg/logging"
	"google.golang.org/grpc"
	"io"
)

var log = logging.GetLogger()

const Service = "atomix.runtime.multimap.v1.CounterMap"

func NewMultiMapServer(protocol *client.Protocol) atomicmultimapv1.MultiMapServer {
	return &multiRaftMultiMapServer{
		Protocol: protocol,
	}
}

type multiRaftMultiMapServer struct {
	*client.Protocol
}

func (s *multiRaftMultiMapServer) Create(ctx context.Context, request *atomicmultimapv1.CreateRequest) (*atomicmultimapv1.CreateResponse, error) {
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
	response := &atomicmultimapv1.CreateResponse{}
	log.Debugw("Create",
		logging.Stringer("CreateRequest", request),
		logging.Stringer("CreateResponse", response))
	return response, nil
}

func (s *multiRaftMultiMapServer) Close(ctx context.Context, request *atomicmultimapv1.CloseRequest) (*atomicmultimapv1.CloseResponse, error) {
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
	response := &atomicmultimapv1.CloseResponse{}
	log.Debugw("Close",
		logging.Stringer("CloseRequest", request),
		logging.Stringer("CloseResponse", response))
	return response, nil
}

func (s *multiRaftMultiMapServer) Size(ctx context.Context, request *atomicmultimapv1.SizeRequest) (*atomicmultimapv1.SizeResponse, error) {
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
			return api.NewMultiMapClient(conn).Size(ctx, &api.SizeRequest{
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
	response := &atomicmultimapv1.SizeResponse{
		Size_: uint32(size),
	}
	log.Debugw("Size",
		logging.Stringer("SizeRequest", request),
		logging.Stringer("SizeResponse", response))
	return response, nil
}

func (s *multiRaftMultiMapServer) Put(ctx context.Context, request *atomicmultimapv1.PutRequest) (*atomicmultimapv1.PutResponse, error) {
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
	_, err = command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (*api.PutResponse, error) {
		input := &api.PutRequest{
			Headers: headers,
			PutInput: &api.PutInput{
				Key:   request.Key,
				Value: request.Value,
			},
		}
		return api.NewMultiMapClient(conn).Put(ctx, input)
	})
	if err != nil {
		log.Warnw("Put",
			logging.Stringer("PutRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &atomicmultimapv1.PutResponse{}
	log.Debugw("Put",
		logging.Stringer("PutRequest", request),
		logging.Stringer("PutResponse", response))
	return response, nil
}

func (s *multiRaftMultiMapServer) PutAll(ctx context.Context, request *atomicmultimapv1.PutAllRequest) (*atomicmultimapv1.PutAllResponse, error) {
	log.Debugw("PutAll",
		logging.Stringer("PutAllRequest", request))
	partition := s.PartitionBy([]byte(request.Key))
	session, err := partition.GetSession(ctx)
	if err != nil {
		log.Warnw("PutAll",
			logging.Stringer("PutAllRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	primitive, err := session.GetPrimitive(request.ID.Name)
	if err != nil {
		log.Warnw("PutAll",
			logging.Stringer("PutAllRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	command := client.Command[*api.PutAllResponse](primitive)
	output, err := command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (*api.PutAllResponse, error) {
		input := &api.PutAllRequest{
			Headers: headers,
			PutAllInput: &api.PutAllInput{
				Key:    request.Key,
				Values: request.Values,
			},
		}
		return api.NewMultiMapClient(conn).PutAll(ctx, input)
	})
	if err != nil {
		log.Warnw("PutAll",
			logging.Stringer("PutAllRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &atomicmultimapv1.PutAllResponse{
		Updated: output.Updated,
	}
	log.Debugw("PutAll",
		logging.Stringer("PutAllRequest", request),
		logging.Stringer("PutAllResponse", response))
	return response, nil
}

func (s *multiRaftMultiMapServer) Replace(ctx context.Context, request *atomicmultimapv1.ReplaceRequest) (*atomicmultimapv1.ReplaceResponse, error) {
	log.Debugw("Replace",
		logging.Stringer("ReplaceRequest", request))
	partition := s.PartitionBy([]byte(request.Key))
	session, err := partition.GetSession(ctx)
	if err != nil {
		log.Warnw("Replace",
			logging.Stringer("ReplaceRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	primitive, err := session.GetPrimitive(request.ID.Name)
	if err != nil {
		log.Warnw("Replace",
			logging.Stringer("ReplaceRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	command := client.Command[*api.ReplaceResponse](primitive)
	output, err := command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (*api.ReplaceResponse, error) {
		input := &api.ReplaceRequest{
			Headers: headers,
			ReplaceInput: &api.ReplaceInput{
				Key:    request.Key,
				Values: request.Values,
			},
		}
		return api.NewMultiMapClient(conn).Replace(ctx, input)
	})
	if err != nil {
		log.Warnw("Replace",
			logging.Stringer("ReplaceRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &atomicmultimapv1.ReplaceResponse{
		PrevValues: output.PrevValues,
	}
	log.Debugw("Replace",
		logging.Stringer("ReplaceRequest", request),
		logging.Stringer("ReplaceResponse", response))
	return response, nil
}

func (s *multiRaftMultiMapServer) Contains(ctx context.Context, request *atomicmultimapv1.ContainsRequest) (*atomicmultimapv1.ContainsResponse, error) {
	log.Debugw("Contains",
		logging.Stringer("ContainsRequest", request))
	partition := s.PartitionBy([]byte(request.Key))
	session, err := partition.GetSession(ctx)
	if err != nil {
		log.Warnw("Contains",
			logging.Stringer("ContainsRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	primitive, err := session.GetPrimitive(request.ID.Name)
	if err != nil {
		log.Warnw("Contains",
			logging.Stringer("ContainsRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	query := client.Query[*api.ContainsResponse](primitive)
	output, err := query.Run(func(conn *grpc.ClientConn, headers *multiraftv1.QueryRequestHeaders) (*api.ContainsResponse, error) {
		return api.NewMultiMapClient(conn).Contains(ctx, &api.ContainsRequest{
			Headers: headers,
			ContainsInput: &api.ContainsInput{
				Key: request.Key,
			},
		})
	})
	if err != nil {
		log.Warnw("Contains",
			logging.Stringer("ContainsRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &atomicmultimapv1.ContainsResponse{
		Result: output.Result,
	}
	log.Debugw("Contains",
		logging.Stringer("ContainsRequest", request),
		logging.Stringer("ContainsResponse", response))
	return response, nil
}

func (s *multiRaftMultiMapServer) Get(ctx context.Context, request *atomicmultimapv1.GetRequest) (*atomicmultimapv1.GetResponse, error) {
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
		return api.NewMultiMapClient(conn).Get(ctx, &api.GetRequest{
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
	response := &atomicmultimapv1.GetResponse{
		Values: output.Values,
	}
	log.Debugw("Get",
		logging.Stringer("GetRequest", request),
		logging.Stringer("GetResponse", response))
	return response, nil
}

func (s *multiRaftMultiMapServer) Remove(ctx context.Context, request *atomicmultimapv1.RemoveRequest) (*atomicmultimapv1.RemoveResponse, error) {
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
	_, err = command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (*api.RemoveResponse, error) {
		input := &api.RemoveRequest{
			Headers: headers,
			RemoveInput: &api.RemoveInput{
				Key:   request.Key,
				Value: request.Value,
			},
		}
		return api.NewMultiMapClient(conn).Remove(ctx, input)
	})
	if err != nil {
		log.Warnw("Remove",
			logging.Stringer("RemoveRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &atomicmultimapv1.RemoveResponse{}
	log.Debugw("Remove",
		logging.Stringer("RemoveRequest", request),
		logging.Stringer("RemoveResponse", response))
	return response, nil
}

func (s *multiRaftMultiMapServer) RemoveAll(ctx context.Context, request *atomicmultimapv1.RemoveAllRequest) (*atomicmultimapv1.RemoveAllResponse, error) {
	log.Debugw("RemoveAll",
		logging.Stringer("RemoveAllRequest", request))
	partition := s.PartitionBy([]byte(request.Key))
	session, err := partition.GetSession(ctx)
	if err != nil {
		log.Warnw("RemoveAll",
			logging.Stringer("RemoveAllRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	primitive, err := session.GetPrimitive(request.ID.Name)
	if err != nil {
		log.Warnw("RemoveAll",
			logging.Stringer("RemoveAllRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	command := client.Command[*api.RemoveAllResponse](primitive)
	output, err := command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (*api.RemoveAllResponse, error) {
		input := &api.RemoveAllRequest{
			Headers: headers,
			RemoveAllInput: &api.RemoveAllInput{
				Key: request.Key,
			},
		}
		return api.NewMultiMapClient(conn).RemoveAll(ctx, input)
	})
	if err != nil {
		log.Warnw("RemoveAll",
			logging.Stringer("RemoveAllRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &atomicmultimapv1.RemoveAllResponse{
		Values: output.Values,
	}
	log.Debugw("RemoveAll",
		logging.Stringer("RemoveAllRequest", request),
		logging.Stringer("RemoveAllResponse", response))
	return response, nil
}

func (s *multiRaftMultiMapServer) Clear(ctx context.Context, request *atomicmultimapv1.ClearRequest) (*atomicmultimapv1.ClearResponse, error) {
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
			return api.NewMultiMapClient(conn).Clear(ctx, &api.ClearRequest{
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
	response := &atomicmultimapv1.ClearResponse{}
	log.Debugw("Clear",
		logging.Stringer("ClearRequest", request),
		logging.Stringer("ClearResponse", response))
	return response, nil
}

func (s *multiRaftMultiMapServer) Events(request *atomicmultimapv1.EventsRequest, server atomicmultimapv1.MultiMap_EventsServer) error {
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
			return api.NewMultiMapClient(conn).Events(server.Context(), &api.EventsRequest{
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
			var response *atomicmultimapv1.EventsResponse
			switch e := output.Event.Event.(type) {
			case *api.Event_Inserted_:
				response = &atomicmultimapv1.EventsResponse{
					Event: atomicmultimapv1.Event{
						Key: output.Event.Key,
						Event: &atomicmultimapv1.Event_Inserted_{
							Inserted: &atomicmultimapv1.Event_Inserted{
								Values: e.Inserted.Values,
							},
						},
					},
				}
			case *api.Event_Updated_:
				response = &atomicmultimapv1.EventsResponse{
					Event: atomicmultimapv1.Event{
						Key: output.Event.Key,
						Event: &atomicmultimapv1.Event_Updated_{
							Updated: &atomicmultimapv1.Event_Updated{
								Values:     e.Updated.Values,
								PrevValues: e.Updated.PrevValues,
							},
						},
					},
				}
			case *api.Event_Removed_:
				response = &atomicmultimapv1.EventsResponse{
					Event: atomicmultimapv1.Event{
						Key: output.Event.Key,
						Event: &atomicmultimapv1.Event_Removed_{
							Removed: &atomicmultimapv1.Event_Removed{
								Values: e.Removed.Values,
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

func (s *multiRaftMultiMapServer) Entries(request *atomicmultimapv1.EntriesRequest, server atomicmultimapv1.MultiMap_EntriesServer) error {
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
			return api.NewMultiMapClient(conn).Entries(server.Context(), &api.EntriesRequest{
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
			response := &atomicmultimapv1.EntriesResponse{
				Entry: atomicmultimapv1.Entry{
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

var _ atomicmultimapv1.MultiMapServer = (*multiRaftMultiMapServer)(nil)
