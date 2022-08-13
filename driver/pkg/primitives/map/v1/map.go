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
	mapv1 "github.com/atomix/runtime/api/atomix/runtime/map/v1"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/atomix/runtime/sdk/pkg/logging"
	"github.com/atomix/runtime/sdk/pkg/runtime"
	"google.golang.org/grpc"
	"io"
)

var log = logging.GetLogger()

const Service = "atomix.multiraft.map.v1.Map"

func NewMapServer(protocol *client.Protocol, config api.MapConfig) mapv1.MapServer {
	return &MapServer{
		Protocol: protocol,
	}
}

type MapServer struct {
	*client.Protocol
}

func (s *MapServer) Create(ctx context.Context, request *mapv1.CreateRequest) (*mapv1.CreateResponse, error) {
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
	response := &mapv1.CreateResponse{}
	log.Debugw("Create",
		logging.Stringer("CreateRequest", request),
		logging.Stringer("CreateResponse", response))
	return response, nil
}

func (s *MapServer) Close(ctx context.Context, request *mapv1.CloseRequest) (*mapv1.CloseResponse, error) {
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
	response := &mapv1.CloseResponse{}
	log.Debugw("Close",
		logging.Stringer("CloseRequest", request),
		logging.Stringer("CloseResponse", response))
	return response, nil
}

func (s *MapServer) Size(ctx context.Context, request *mapv1.SizeRequest) (*mapv1.SizeResponse, error) {
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
				Headers:   *headers,
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
	response := &mapv1.SizeResponse{
		Size_: uint32(size),
	}
	log.Debugw("Size",
		logging.Stringer("SizeRequest", request),
		logging.Stringer("SizeResponse", response))
	return response, nil
}

func (s *MapServer) Put(ctx context.Context, request *mapv1.PutRequest) (*mapv1.PutResponse, error) {
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
			Headers: *headers,
			PutInput: &api.PutInput{
				Entry: &api.Entry{
					Key: request.Key,
					Value: api.Value{
						Value: request.Value.Value,
						TTL:   request.Value.TTL,
					},
				},
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
	response := &mapv1.PutResponse{}
	if output.PrevValue != nil {
		response.PrevValue = &mapv1.Value{
			Value: response.PrevValue.Value,
			TTL:   response.PrevValue.TTL,
		}
	}
	log.Debugw("Put",
		logging.Stringer("PutRequest", request),
		logging.Stringer("PutResponse", response))
	return response, nil
}

func (s *MapServer) Get(ctx context.Context, request *mapv1.GetRequest) (*mapv1.GetResponse, error) {
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
			Headers: *headers,
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
	response := &mapv1.GetResponse{
		Value: mapv1.Value{
			Value: output.Value.Value,
			TTL:   output.Value.TTL,
		},
	}
	log.Debugw("Get",
		logging.Stringer("GetRequest", request),
		logging.Stringer("GetResponse", response))
	return response, nil
}

func (s *MapServer) Remove(ctx context.Context, request *mapv1.RemoveRequest) (*mapv1.RemoveResponse, error) {
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
			Headers: *headers,
			RemoveInput: &api.RemoveInput{
				Key: request.Key,
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
	response := &mapv1.RemoveResponse{}
	if output.Value != nil {
		response.Value = &mapv1.Value{
			Value: response.Value.Value,
			TTL:   response.Value.TTL,
		}
	}
	log.Debugw("Remove",
		logging.Stringer("RemoveRequest", request),
		logging.Stringer("RemoveResponse", response))
	return response, nil
}

func (s *MapServer) Clear(ctx context.Context, request *mapv1.ClearRequest) (*mapv1.ClearResponse, error) {
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
				Headers:    *headers,
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
	response := &mapv1.ClearResponse{}
	log.Debugw("Clear",
		logging.Stringer("ClearRequest", request),
		logging.Stringer("ClearResponse", response))
	return response, nil
}

func (s *MapServer) Events(request *mapv1.EventsRequest, server mapv1.Map_EventsServer) error {
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
		command := client.StreamCommand[api.Map_EventsClient, *api.EventsResponse](primitive)
		stream, err := command.Open(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (api.Map_EventsClient, error) {
			return api.NewMapClient(conn).Events(server.Context(), &api.EventsRequest{
				Headers: *headers,
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
			var response *mapv1.EventsResponse
			switch e := output.Event.Event.(type) {
			case *api.Event_Inserted_:
				response = &mapv1.EventsResponse{
					Event: mapv1.Event{
						Key: output.Event.Key,
						Event: &mapv1.Event_Inserted_{
							Inserted: &mapv1.Event_Inserted{
								Value: mapv1.Value{
									Value: e.Inserted.Value.Value,
									TTL:   e.Inserted.Value.TTL,
								},
							},
						},
					},
				}
			case *api.Event_Updated_:
				response = &mapv1.EventsResponse{
					Event: mapv1.Event{
						Key: output.Event.Key,
						Event: &mapv1.Event_Updated_{
							Updated: &mapv1.Event_Updated{
								Value: mapv1.Value{
									Value: e.Updated.NewValue.Value,
									TTL:   e.Updated.NewValue.TTL,
								},
								PrevValue: mapv1.Value{
									Value: e.Updated.PrevValue.Value,
									TTL:   e.Updated.PrevValue.TTL,
								},
							},
						},
					},
				}
			case *api.Event_Removed_:
				response = &mapv1.EventsResponse{
					Event: mapv1.Event{
						Key: output.Event.Key,
						Event: &mapv1.Event_Removed_{
							Removed: &mapv1.Event_Removed{
								Value: mapv1.Value{
									Value: e.Removed.Value.Value,
									TTL:   e.Removed.Value.TTL,
								},
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

func (s *MapServer) Entries(request *mapv1.EntriesRequest, server mapv1.Map_EntriesServer) error {
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
		query := client.StreamQuery[api.Map_EntriesClient, *api.EntriesResponse](primitive)
		stream, err := query.Open(func(conn *grpc.ClientConn, headers *multiraftv1.QueryRequestHeaders) (api.Map_EntriesClient, error) {
			return api.NewMapClient(conn).Entries(server.Context(), &api.EntriesRequest{
				Headers: *headers,
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
			response := &mapv1.EntriesResponse{
				Entry: mapv1.Entry{
					Key: output.Entry.Key,
					Value: &mapv1.Value{
						Value: output.Entry.Value.Value,
						TTL:   output.Entry.Value.TTL,
					},
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

var _ mapv1.MapServer = (*MapServer)(nil)
