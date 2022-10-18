// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"context"
	api "github.com/atomix/multi-raft-storage/api/atomix/multiraft/set/v1"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/driver/pkg/client"
	"github.com/atomix/multi-raft-storage/driver/pkg/util/async"
	atomicsetv1 "github.com/atomix/runtime/api/atomix/runtime/set/v1"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/atomix/runtime/sdk/pkg/logging"
	"github.com/atomix/runtime/sdk/pkg/runtime"
	"github.com/atomix/runtime/sdk/pkg/stringer"
	"google.golang.org/grpc"
	"io"
)

var log = logging.GetLogger()

const Service = "atomix.runtime.set.v1.Set"

const truncLen = 200

func NewSetServer(protocol *client.Protocol, spec runtime.PrimitiveSpec) (atomicsetv1.SetServer, error) {
	return &multiRaftSetServer{
		Protocol:      protocol,
		PrimitiveSpec: spec,
	}, nil
}

type multiRaftSetServer struct {
	*client.Protocol
	runtime.PrimitiveSpec
}

func (s *multiRaftSetServer) Create(ctx context.Context, request *atomicsetv1.CreateRequest) (*atomicsetv1.CreateResponse, error) {
	log.Debugw("Create",
		logging.Stringer("CreateRequest", stringer.Truncate(request, truncLen)))
	partitions := s.Partitions()
	err := async.IterAsync(len(partitions), func(i int) error {
		partition := partitions[i]
		session, err := partition.GetSession(ctx)
		if err != nil {
			return err
		}
		return session.CreatePrimitive(ctx, s.PrimitiveSpec)
	})
	if err != nil {
		log.Warnw("Create",
			logging.Stringer("CreateRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &atomicsetv1.CreateResponse{}
	log.Debugw("Create",
		logging.Stringer("CreateRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("CreateResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *multiRaftSetServer) Close(ctx context.Context, request *atomicsetv1.CloseRequest) (*atomicsetv1.CloseResponse, error) {
	log.Debugw("Close",
		logging.Stringer("CloseRequest", stringer.Truncate(request, truncLen)))
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
			logging.Stringer("CloseRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &atomicsetv1.CloseResponse{}
	log.Debugw("Close",
		logging.Stringer("CloseRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("CloseResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *multiRaftSetServer) Size(ctx context.Context, request *atomicsetv1.SizeRequest) (*atomicsetv1.SizeResponse, error) {
	log.Debugw("Size",
		logging.Stringer("SizeRequest", stringer.Truncate(request, truncLen)))
	partitions := s.Partitions()
	sizes, err := async.ExecuteAsync[int](len(partitions), func(i int) (int, error) {
		partition := partitions[i]
		session, err := partition.GetSession(ctx)
		if err != nil {
			log.Warnw("Size",
				logging.Stringer("SizeRequest", stringer.Truncate(request, truncLen)),
				logging.Error("Error", err))
			return 0, err
		}
		primitive, err := session.GetPrimitive(request.ID.Name)
		if err != nil {
			log.Warnw("Size",
				logging.Stringer("SizeRequest", stringer.Truncate(request, truncLen)),
				logging.Error("Error", err))
			return 0, err
		}
		query := client.Query[*api.SizeResponse](primitive)
		output, err := query.Run(func(conn *grpc.ClientConn, headers *multiraftv1.QueryRequestHeaders) (*api.SizeResponse, error) {
			return api.NewSetClient(conn).Size(ctx, &api.SizeRequest{
				Headers:   headers,
				SizeInput: &api.SizeInput{},
			})
		})
		if err != nil {
			log.Warnw("Size",
				logging.Stringer("SizeRequest", stringer.Truncate(request, truncLen)),
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
	response := &atomicsetv1.SizeResponse{
		Size_: uint32(size),
	}
	log.Debugw("Size",
		logging.Stringer("SizeRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("SizeResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *multiRaftSetServer) Add(ctx context.Context, request *atomicsetv1.AddRequest) (*atomicsetv1.AddResponse, error) {
	log.Debugw("Add",
		logging.Stringer("AddRequest", stringer.Truncate(request, truncLen)))
	partition := s.PartitionBy([]byte(request.Element.Value))
	session, err := partition.GetSession(ctx)
	if err != nil {
		log.Warnw("Add",
			logging.Stringer("AddRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	primitive, err := session.GetPrimitive(request.ID.Name)
	if err != nil {
		log.Warnw("Add",
			logging.Stringer("AddRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	command := client.Command[*api.AddResponse](primitive)
	_, err = command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (*api.AddResponse, error) {
		input := &api.AddRequest{
			Headers: headers,
			AddInput: &api.AddInput{
				Element: api.Element{
					Value: request.Element.Value,
				},
				TTL: request.TTL,
			},
		}
		return api.NewSetClient(conn).Add(ctx, input)
	})
	if err != nil {
		log.Warnw("Add",
			logging.Stringer("AddRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &atomicsetv1.AddResponse{}
	log.Debugw("Add",
		logging.Stringer("AddRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("AddResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *multiRaftSetServer) Contains(ctx context.Context, request *atomicsetv1.ContainsRequest) (*atomicsetv1.ContainsResponse, error) {
	log.Debugw("Contains",
		logging.Stringer("ContainsRequest", stringer.Truncate(request, truncLen)))
	partition := s.PartitionBy([]byte(request.Element.Value))
	session, err := partition.GetSession(ctx)
	if err != nil {
		log.Warnw("Contains",
			logging.Stringer("ContainsRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	primitive, err := session.GetPrimitive(request.ID.Name)
	if err != nil {
		log.Warnw("Contains",
			logging.Stringer("ContainsRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	command := client.Query[*api.ContainsResponse](primitive)
	output, err := command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.QueryRequestHeaders) (*api.ContainsResponse, error) {
		input := &api.ContainsRequest{
			Headers: headers,
			ContainsInput: &api.ContainsInput{
				Element: api.Element{
					Value: request.Element.Value,
				},
			},
		}
		return api.NewSetClient(conn).Contains(ctx, input)
	})
	if err != nil {
		log.Warnw("Contains",
			logging.Stringer("ContainsRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &atomicsetv1.ContainsResponse{
		Contains: output.Contains,
	}
	log.Debugw("Contains",
		logging.Stringer("ContainsRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("ContainsResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *multiRaftSetServer) Remove(ctx context.Context, request *atomicsetv1.RemoveRequest) (*atomicsetv1.RemoveResponse, error) {
	log.Debugw("Remove",
		logging.Stringer("RemoveRequest", stringer.Truncate(request, truncLen)))
	partition := s.PartitionBy([]byte(request.Element.Value))
	session, err := partition.GetSession(ctx)
	if err != nil {
		log.Warnw("Remove",
			logging.Stringer("RemoveRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	primitive, err := session.GetPrimitive(request.ID.Name)
	if err != nil {
		log.Warnw("Remove",
			logging.Stringer("RemoveRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	command := client.Command[*api.RemoveResponse](primitive)
	_, err = command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (*api.RemoveResponse, error) {
		input := &api.RemoveRequest{
			Headers: headers,
			RemoveInput: &api.RemoveInput{
				Element: api.Element{
					Value: request.Element.Value,
				},
			},
		}
		return api.NewSetClient(conn).Remove(ctx, input)
	})
	if err != nil {
		log.Warnw("Remove",
			logging.Stringer("RemoveRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &atomicsetv1.RemoveResponse{}
	log.Debugw("Remove",
		logging.Stringer("RemoveRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("RemoveResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *multiRaftSetServer) Clear(ctx context.Context, request *atomicsetv1.ClearRequest) (*atomicsetv1.ClearResponse, error) {
	log.Debugw("Clear",
		logging.Stringer("ClearRequest", stringer.Truncate(request, truncLen)))
	partitions := s.Partitions()
	err := async.IterAsync(len(partitions), func(i int) error {
		partition := partitions[i]
		session, err := partition.GetSession(ctx)
		if err != nil {
			log.Warnw("Clear",
				logging.Stringer("ClearRequest", stringer.Truncate(request, truncLen)),
				logging.Error("Error", err))
			return err
		}
		primitive, err := session.GetPrimitive(request.ID.Name)
		if err != nil {
			log.Warnw("Clear",
				logging.Stringer("ClearRequest", stringer.Truncate(request, truncLen)),
				logging.Error("Error", err))
			return err
		}
		command := client.Command[*api.ClearResponse](primitive)
		_, err = command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (*api.ClearResponse, error) {
			return api.NewSetClient(conn).Clear(ctx, &api.ClearRequest{
				Headers:    headers,
				ClearInput: &api.ClearInput{},
			})
		})
		if err != nil {
			log.Warnw("Clear",
				logging.Stringer("ClearRequest", stringer.Truncate(request, truncLen)),
				logging.Error("Error", err))
			return err
		}
		return nil
	})
	if err != nil {
		return nil, errors.ToProto(err)
	}
	response := &atomicsetv1.ClearResponse{}
	log.Debugw("Clear",
		logging.Stringer("ClearRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("ClearResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *multiRaftSetServer) Events(request *atomicsetv1.EventsRequest, server atomicsetv1.Set_EventsServer) error {
	log.Debugw("Events",
		logging.Stringer("EventsRequest", stringer.Truncate(request, truncLen)))
	partitions := s.Partitions()
	return async.IterAsync(len(partitions), func(i int) error {
		partition := partitions[i]
		session, err := partition.GetSession(server.Context())
		if err != nil {
			err = errors.ToProto(err)
			log.Warnw("Events",
				logging.Stringer("EventsRequest", stringer.Truncate(request, truncLen)),
				logging.Error("Error", err))
			return err
		}
		primitive, err := session.GetPrimitive(request.ID.Name)
		if err != nil {
			err = errors.ToProto(err)
			log.Warnw("Events",
				logging.Stringer("EventsRequest", stringer.Truncate(request, truncLen)),
				logging.Error("Error", err))
			return err
		}
		command := client.StreamCommand[*api.EventsResponse](primitive)
		stream, err := command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (client.CommandStream[*api.EventsResponse], error) {
			return api.NewSetClient(conn).Events(server.Context(), &api.EventsRequest{
				Headers:     headers,
				EventsInput: &api.EventsInput{},
			})
		})
		if err != nil {
			err = errors.ToProto(err)
			log.Warnw("Events",
				logging.Stringer("EventsRequest", stringer.Truncate(request, truncLen)),
				logging.Error("Error", err))
			return err
		}
		for {
			output, err := stream.Recv()
			if err == io.EOF {
				log.Debugw("Events",
					logging.Stringer("EventsRequest", stringer.Truncate(request, truncLen)),
					logging.String("State", "Done"))
				return nil
			}
			if err != nil {
				log.Warnw("Events",
					logging.Stringer("EventsRequest", stringer.Truncate(request, truncLen)),
					logging.Error("Error", err))
				return errors.ToProto(err)
			}
			var response *atomicsetv1.EventsResponse
			switch e := output.Event.Event.(type) {
			case *api.Event_Added_:
				response = &atomicsetv1.EventsResponse{
					Event: atomicsetv1.Event{
						Event: &atomicsetv1.Event_Added_{
							Added: &atomicsetv1.Event_Added{
								Element: atomicsetv1.Element{
									Value: e.Added.Element.Value,
								},
							},
						},
					},
				}
			case *api.Event_Removed_:
				response = &atomicsetv1.EventsResponse{
					Event: atomicsetv1.Event{
						Event: &atomicsetv1.Event_Removed_{
							Removed: &atomicsetv1.Event_Removed{
								Element: atomicsetv1.Element{
									Value: e.Removed.Element.Value,
								},
								Expired: e.Removed.Expired,
							},
						},
					},
				}
			}
			log.Debugw("Events",
				logging.Stringer("EventsRequest", stringer.Truncate(request, truncLen)),
				logging.Stringer("EventsResponse", stringer.Truncate(response, truncLen)))
			if err := server.Send(response); err != nil {
				log.Warnw("Events",
					logging.Stringer("EventsRequest", stringer.Truncate(request, truncLen)),
					logging.Stringer("EventsResponse", stringer.Truncate(response, truncLen)),
					logging.Error("Error", err))
				return err
			}
		}
	})
}

func (s *multiRaftSetServer) Elements(request *atomicsetv1.ElementsRequest, server atomicsetv1.Set_ElementsServer) error {
	log.Debugw("Elements",
		logging.Stringer("ElementsRequest", stringer.Truncate(request, truncLen)))
	partitions := s.Partitions()
	return async.IterAsync(len(partitions), func(i int) error {
		partition := partitions[i]
		session, err := partition.GetSession(server.Context())
		if err != nil {
			log.Warnw("Elements",
				logging.Stringer("ElementsRequest", stringer.Truncate(request, truncLen)),
				logging.Error("Error", err))
			return errors.ToProto(err)
		}
		primitive, err := session.GetPrimitive(request.ID.Name)
		if err != nil {
			log.Warnw("Elements",
				logging.Stringer("ElementsRequest", stringer.Truncate(request, truncLen)),
				logging.Error("Error", err))
			return errors.ToProto(err)
		}
		query := client.StreamQuery[*api.ElementsResponse](primitive)
		stream, err := query.Run(func(conn *grpc.ClientConn, headers *multiraftv1.QueryRequestHeaders) (client.QueryStream[*api.ElementsResponse], error) {
			return api.NewSetClient(conn).Elements(server.Context(), &api.ElementsRequest{
				Headers: headers,
				ElementsInput: &api.ElementsInput{
					Watch: request.Watch,
				},
			})
		})
		if err != nil {
			log.Warnw("Elements",
				logging.Stringer("ElementsRequest", stringer.Truncate(request, truncLen)),
				logging.Error("Error", err))
			return errors.ToProto(err)
		}
		for {
			output, err := stream.Recv()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				log.Warnw("Elements",
					logging.Stringer("ElementsRequest", stringer.Truncate(request, truncLen)),
					logging.Error("Error", err))
				return errors.ToProto(err)
			}
			response := &atomicsetv1.ElementsResponse{
				Element: atomicsetv1.Element{
					Value: output.Element.Value,
				},
			}
			log.Debugw("Elements",
				logging.Stringer("ElementsRequest", stringer.Truncate(request, truncLen)),
				logging.Stringer("ElementsResponse", stringer.Truncate(response, truncLen)))
			if err := server.Send(response); err != nil {
				log.Warnw("Elements",
					logging.Stringer("ElementsRequest", stringer.Truncate(request, truncLen)),
					logging.Stringer("ElementsResponse", stringer.Truncate(response, truncLen)),
					logging.Error("Error", err))
				return err
			}
		}
	})
}

var _ atomicsetv1.SetServer = (*multiRaftSetServer)(nil)
