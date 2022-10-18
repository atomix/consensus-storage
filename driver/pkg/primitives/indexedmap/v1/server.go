// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"context"
	api "github.com/atomix/multi-raft-storage/api/atomix/multiraft/indexedmap/v1"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/driver/pkg/client"
	indexedmapv1 "github.com/atomix/runtime/api/atomix/runtime/indexedmap/v1"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/atomix/runtime/sdk/pkg/logging"
	"github.com/atomix/runtime/sdk/pkg/stringer"
	"google.golang.org/grpc"
	"io"
)

var log = logging.GetLogger()

const Service = "atomix.runtime.indexedmap.v1.IndexedMap"

const truncLen = 200

func NewIndexedMapServer(protocol *client.Protocol, config api.IndexedMapConfig) indexedmapv1.IndexedMapServer {
	return &multiRaftIndexedMapServer{
		Protocol: protocol,
	}
}

type multiRaftIndexedMapServer struct {
	*client.Protocol
}

func (s *multiRaftIndexedMapServer) Create(ctx context.Context, request *indexedmapv1.CreateRequest) (*indexedmapv1.CreateResponse, error) {
	log.Debugw("Create",
		logging.Stringer("CreateRequest", stringer.Truncate(request, truncLen)))
	partition := s.PartitionBy([]byte(request.ID.Name))
	session, err := partition.GetSession(ctx)
	if err != nil {
		log.Warnw("Create",
			logging.Stringer("CreateRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	if err := session.CreatePrimitive(ctx, request.ID.Name, Service); err != nil {
		log.Warnw("Create",
			logging.Stringer("CreateRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &indexedmapv1.CreateResponse{}
	log.Debugw("Create",
		logging.Stringer("CreateRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("CreateResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *multiRaftIndexedMapServer) Close(ctx context.Context, request *indexedmapv1.CloseRequest) (*indexedmapv1.CloseResponse, error) {
	log.Debugw("Close",
		logging.Stringer("CloseRequest", stringer.Truncate(request, truncLen)))
	partition := s.PartitionBy([]byte(request.ID.Name))
	session, err := partition.GetSession(ctx)
	if err != nil {
		log.Warnw("Close",
			logging.Stringer("CloseRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	if err := session.ClosePrimitive(ctx, request.ID.Name); err != nil {
		log.Warnw("Close",
			logging.Stringer("CloseRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &indexedmapv1.CloseResponse{}
	log.Debugw("Close",
		logging.Stringer("CloseRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("CloseResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *multiRaftIndexedMapServer) Size(ctx context.Context, request *indexedmapv1.SizeRequest) (*indexedmapv1.SizeResponse, error) {
	log.Debugw("Size",
		logging.Stringer("SizeRequest", stringer.Truncate(request, truncLen)))
	partition := s.PartitionBy([]byte(request.ID.Name))
	session, err := partition.GetSession(ctx)
	if err != nil {
		log.Warnw("Size",
			logging.Stringer("SizeRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	primitive, err := session.GetPrimitive(request.ID.Name)
	if err != nil {
		log.Warnw("Size",
			logging.Stringer("SizeRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	query := client.Query[*api.SizeResponse](primitive)
	output, err := query.Run(func(conn *grpc.ClientConn, headers *multiraftv1.QueryRequestHeaders) (*api.SizeResponse, error) {
		return api.NewIndexedMapClient(conn).Size(ctx, &api.SizeRequest{
			Headers:   headers,
			SizeInput: &api.SizeInput{},
		})
	})
	if err != nil {
		log.Warnw("Size",
			logging.Stringer("SizeRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &indexedmapv1.SizeResponse{
		Size_: output.Size_,
	}
	log.Debugw("Size",
		logging.Stringer("SizeRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("SizeResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *multiRaftIndexedMapServer) Append(ctx context.Context, request *indexedmapv1.AppendRequest) (*indexedmapv1.AppendResponse, error) {
	log.Debugw("Append",
		logging.Stringer("AppendRequest", stringer.Truncate(request, truncLen)))
	partition := s.PartitionBy([]byte(request.ID.Name))
	session, err := partition.GetSession(ctx)
	if err != nil {
		log.Warnw("Append",
			logging.Stringer("AppendRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	primitive, err := session.GetPrimitive(request.ID.Name)
	if err != nil {
		log.Warnw("Append",
			logging.Stringer("AppendRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	command := client.Command[*api.AppendResponse](primitive)
	output, err := command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (*api.AppendResponse, error) {
		return api.NewIndexedMapClient(conn).Append(ctx, &api.AppendRequest{
			Headers: headers,
			AppendInput: &api.AppendInput{
				Key:   request.Key,
				Value: request.Value,
				TTL:   request.TTL,
			},
		})
	})
	if err != nil {
		log.Warnw("Append",
			logging.Stringer("AppendRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &indexedmapv1.AppendResponse{
		Entry: newEntry(output.Entry),
	}
	log.Debugw("Append",
		logging.Stringer("AppendRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("AppendResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *multiRaftIndexedMapServer) Update(ctx context.Context, request *indexedmapv1.UpdateRequest) (*indexedmapv1.UpdateResponse, error) {
	log.Debugw("Update",
		logging.Stringer("UpdateRequest", stringer.Truncate(request, truncLen)))
	partition := s.PartitionBy([]byte(request.ID.Name))
	session, err := partition.GetSession(ctx)
	if err != nil {
		log.Warnw("Update",
			logging.Stringer("UpdateRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	primitive, err := session.GetPrimitive(request.ID.Name)
	if err != nil {
		log.Warnw("Update",
			logging.Stringer("UpdateRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	command := client.Command[*api.UpdateResponse](primitive)
	output, err := command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (*api.UpdateResponse, error) {
		return api.NewIndexedMapClient(conn).Update(ctx, &api.UpdateRequest{
			Headers: headers,
			UpdateInput: &api.UpdateInput{
				Key:         request.Key,
				Index:       request.Index,
				Value:       request.Value,
				TTL:         request.TTL,
				PrevVersion: request.PrevVersion,
			},
		})
	})
	if err != nil {
		log.Warnw("Update",
			logging.Stringer("UpdateRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &indexedmapv1.UpdateResponse{
		Entry: newEntry(output.Entry),
	}
	log.Debugw("Update",
		logging.Stringer("UpdateRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("UpdateResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *multiRaftIndexedMapServer) Get(ctx context.Context, request *indexedmapv1.GetRequest) (*indexedmapv1.GetResponse, error) {
	log.Debugw("Get",
		logging.Stringer("GetRequest", stringer.Truncate(request, truncLen)))
	partition := s.PartitionBy([]byte(request.ID.Name))
	session, err := partition.GetSession(ctx)
	if err != nil {
		log.Warnw("Get",
			logging.Stringer("GetRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	primitive, err := session.GetPrimitive(request.ID.Name)
	if err != nil {
		log.Warnw("Get",
			logging.Stringer("GetRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	query := client.Query[*api.GetResponse](primitive)
	output, err := query.Run(func(conn *grpc.ClientConn, headers *multiraftv1.QueryRequestHeaders) (*api.GetResponse, error) {
		return api.NewIndexedMapClient(conn).Get(ctx, &api.GetRequest{
			Headers: headers,
			GetInput: &api.GetInput{
				Key:   request.Key,
				Index: request.Index,
			},
		})
	})
	if err != nil {
		log.Warnw("Get",
			logging.Stringer("GetRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &indexedmapv1.GetResponse{
		Entry: newEntry(output.Entry),
	}
	log.Debugw("Get",
		logging.Stringer("GetRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("GetResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *multiRaftIndexedMapServer) FirstEntry(ctx context.Context, request *indexedmapv1.FirstEntryRequest) (*indexedmapv1.FirstEntryResponse, error) {
	log.Debugw("FirstEntry",
		logging.Stringer("FirstEntryRequest", stringer.Truncate(request, truncLen)))
	partition := s.PartitionBy([]byte(request.ID.Name))
	session, err := partition.GetSession(ctx)
	if err != nil {
		log.Warnw("FirstEntry",
			logging.Stringer("FirstEntryRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	primitive, err := session.GetPrimitive(request.ID.Name)
	if err != nil {
		log.Warnw("FirstEntry",
			logging.Stringer("FirstEntryRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	query := client.Query[*api.FirstEntryResponse](primitive)
	output, err := query.Run(func(conn *grpc.ClientConn, headers *multiraftv1.QueryRequestHeaders) (*api.FirstEntryResponse, error) {
		return api.NewIndexedMapClient(conn).FirstEntry(ctx, &api.FirstEntryRequest{
			Headers:         headers,
			FirstEntryInput: &api.FirstEntryInput{},
		})
	})
	if err != nil {
		log.Warnw("FirstEntry",
			logging.Stringer("FirstEntryRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &indexedmapv1.FirstEntryResponse{
		Entry: newEntry(output.Entry),
	}
	log.Debugw("FirstEntry",
		logging.Stringer("FirstEntryRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("FirstEntryResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *multiRaftIndexedMapServer) LastEntry(ctx context.Context, request *indexedmapv1.LastEntryRequest) (*indexedmapv1.LastEntryResponse, error) {
	log.Debugw("LastEntry",
		logging.Stringer("LastEntryRequest", stringer.Truncate(request, truncLen)))
	partition := s.PartitionBy([]byte(request.ID.Name))
	session, err := partition.GetSession(ctx)
	if err != nil {
		log.Warnw("LastEntry",
			logging.Stringer("LastEntryRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	primitive, err := session.GetPrimitive(request.ID.Name)
	if err != nil {
		log.Warnw("LastEntry",
			logging.Stringer("LastEntryRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	query := client.Query[*api.LastEntryResponse](primitive)
	output, err := query.Run(func(conn *grpc.ClientConn, headers *multiraftv1.QueryRequestHeaders) (*api.LastEntryResponse, error) {
		return api.NewIndexedMapClient(conn).LastEntry(ctx, &api.LastEntryRequest{
			Headers:        headers,
			LastEntryInput: &api.LastEntryInput{},
		})
	})
	if err != nil {
		log.Warnw("LastEntry",
			logging.Stringer("LastEntryRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &indexedmapv1.LastEntryResponse{
		Entry: newEntry(output.Entry),
	}
	log.Debugw("LastEntry",
		logging.Stringer("LastEntryRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("LastEntryResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *multiRaftIndexedMapServer) NextEntry(ctx context.Context, request *indexedmapv1.NextEntryRequest) (*indexedmapv1.NextEntryResponse, error) {
	log.Debugw("NextEntry",
		logging.Stringer("NextEntryRequest", stringer.Truncate(request, truncLen)))
	partition := s.PartitionBy([]byte(request.ID.Name))
	session, err := partition.GetSession(ctx)
	if err != nil {
		log.Warnw("NextEntry",
			logging.Stringer("NextEntryRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	primitive, err := session.GetPrimitive(request.ID.Name)
	if err != nil {
		log.Warnw("NextEntry",
			logging.Stringer("NextEntryRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	query := client.Query[*api.NextEntryResponse](primitive)
	output, err := query.Run(func(conn *grpc.ClientConn, headers *multiraftv1.QueryRequestHeaders) (*api.NextEntryResponse, error) {
		return api.NewIndexedMapClient(conn).NextEntry(ctx, &api.NextEntryRequest{
			Headers: headers,
			NextEntryInput: &api.NextEntryInput{
				Index: request.Index,
			},
		})
	})
	if err != nil {
		log.Warnw("NextEntry",
			logging.Stringer("NextEntryRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &indexedmapv1.NextEntryResponse{
		Entry: newEntry(output.Entry),
	}
	log.Debugw("NextEntry",
		logging.Stringer("NextEntryRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("NextEntryResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *multiRaftIndexedMapServer) PrevEntry(ctx context.Context, request *indexedmapv1.PrevEntryRequest) (*indexedmapv1.PrevEntryResponse, error) {
	log.Debugw("PrevEntry",
		logging.Stringer("PrevEntryRequest", stringer.Truncate(request, truncLen)))
	partition := s.PartitionBy([]byte(request.ID.Name))
	session, err := partition.GetSession(ctx)
	if err != nil {
		log.Warnw("PrevEntry",
			logging.Stringer("PrevEntryRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	primitive, err := session.GetPrimitive(request.ID.Name)
	if err != nil {
		log.Warnw("PrevEntry",
			logging.Stringer("PrevEntryRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	query := client.Query[*api.PrevEntryResponse](primitive)
	output, err := query.Run(func(conn *grpc.ClientConn, headers *multiraftv1.QueryRequestHeaders) (*api.PrevEntryResponse, error) {
		return api.NewIndexedMapClient(conn).PrevEntry(ctx, &api.PrevEntryRequest{
			Headers: headers,
			PrevEntryInput: &api.PrevEntryInput{
				Index: request.Index,
			},
		})
	})
	if err != nil {
		log.Warnw("PrevEntry",
			logging.Stringer("PrevEntryRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &indexedmapv1.PrevEntryResponse{
		Entry: newEntry(output.Entry),
	}
	log.Debugw("PrevEntry",
		logging.Stringer("PrevEntryRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("PrevEntryResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *multiRaftIndexedMapServer) Remove(ctx context.Context, request *indexedmapv1.RemoveRequest) (*indexedmapv1.RemoveResponse, error) {
	log.Debugw("Remove",
		logging.Stringer("RemoveRequest", stringer.Truncate(request, truncLen)))
	partition := s.PartitionBy([]byte(request.ID.Name))
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
	output, err := command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (*api.RemoveResponse, error) {
		return api.NewIndexedMapClient(conn).Remove(ctx, &api.RemoveRequest{
			Headers: headers,
			RemoveInput: &api.RemoveInput{
				Key:         request.Key,
				Index:       request.Index,
				PrevVersion: request.PrevVersion,
			},
		})
	})
	if err != nil {
		log.Warnw("Remove",
			logging.Stringer("RemoveRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &indexedmapv1.RemoveResponse{
		Entry: newEntry(output.Entry),
	}
	log.Debugw("Remove",
		logging.Stringer("RemoveRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("RemoveResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *multiRaftIndexedMapServer) Clear(ctx context.Context, request *indexedmapv1.ClearRequest) (*indexedmapv1.ClearResponse, error) {
	log.Debugw("Clear",
		logging.Stringer("ClearRequest", stringer.Truncate(request, truncLen)))
	partition := s.PartitionBy([]byte(request.ID.Name))
	session, err := partition.GetSession(ctx)
	if err != nil {
		log.Warnw("Clear",
			logging.Stringer("ClearRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	primitive, err := session.GetPrimitive(request.ID.Name)
	if err != nil {
		log.Warnw("Clear",
			logging.Stringer("ClearRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	command := client.Command[*api.ClearResponse](primitive)
	_, err = command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (*api.ClearResponse, error) {
		return api.NewIndexedMapClient(conn).Clear(ctx, &api.ClearRequest{
			Headers:    headers,
			ClearInput: &api.ClearInput{},
		})
	})
	if err != nil {
		log.Warnw("Clear",
			logging.Stringer("ClearRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &indexedmapv1.ClearResponse{}
	log.Debugw("Clear",
		logging.Stringer("ClearRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("ClearResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *multiRaftIndexedMapServer) Entries(request *indexedmapv1.EntriesRequest, server indexedmapv1.IndexedMap_EntriesServer) error {
	log.Debugw("Entries",
		logging.Stringer("EntriesRequest", stringer.Truncate(request, truncLen)))
	partition := s.PartitionBy([]byte(request.ID.Name))
	session, err := partition.GetSession(server.Context())
	if err != nil {
		log.Warnw("Entries",
			logging.Stringer("EntriesRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return errors.ToProto(err)
	}
	primitive, err := session.GetPrimitive(request.ID.Name)
	if err != nil {
		log.Warnw("Entries",
			logging.Stringer("EntriesRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return errors.ToProto(err)
	}
	query := client.StreamQuery[*api.EntriesResponse](primitive)
	stream, err := query.Run(func(conn *grpc.ClientConn, headers *multiraftv1.QueryRequestHeaders) (client.QueryStream[*api.EntriesResponse], error) {
		return api.NewIndexedMapClient(conn).Entries(server.Context(), &api.EntriesRequest{
			Headers: headers,
			EntriesInput: &api.EntriesInput{
				Watch: request.Watch,
			},
		})
	})
	if err != nil {
		log.Warnw("Entries",
			logging.Stringer("EntriesRequest", stringer.Truncate(request, truncLen)),
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
				logging.Stringer("EntriesRequest", stringer.Truncate(request, truncLen)),
				logging.Error("Error", err))
			return errors.ToProto(err)
		}
		response := &indexedmapv1.EntriesResponse{
			Entry: *newEntry(&output.Entry),
		}
		log.Debugw("Entries",
			logging.Stringer("EntriesRequest", stringer.Truncate(request, truncLen)),
			logging.Stringer("EntriesResponse", stringer.Truncate(response, truncLen)))
		if err := server.Send(response); err != nil {
			log.Warnw("Entries",
				logging.Stringer("EntriesRequest", stringer.Truncate(request, truncLen)),
				logging.Stringer("EntriesResponse", stringer.Truncate(response, truncLen)),
				logging.Error("Error", err))
			return err
		}
	}
}

func (s *multiRaftIndexedMapServer) Events(request *indexedmapv1.EventsRequest, server indexedmapv1.IndexedMap_EventsServer) error {
	log.Debugw("Events",
		logging.Stringer("EventsRequest", stringer.Truncate(request, truncLen)))
	partition := s.PartitionBy([]byte(request.ID.Name))
	session, err := partition.GetSession(server.Context())
	if err != nil {
		log.Warnw("Events",
			logging.Stringer("EventsRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return errors.ToProto(err)
	}
	primitive, err := session.GetPrimitive(request.ID.Name)
	if err != nil {
		log.Warnw("Events",
			logging.Stringer("EventsRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return errors.ToProto(err)
	}
	command := client.StreamCommand[*api.EventsResponse](primitive)
	stream, err := command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (client.CommandStream[*api.EventsResponse], error) {
		return api.NewIndexedMapClient(conn).Events(server.Context(), &api.EventsRequest{
			Headers: headers,
			EventsInput: &api.EventsInput{
				Key: request.Key,
			},
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
		var response *indexedmapv1.EventsResponse
		switch e := output.Event.Event.(type) {
		case *api.Event_Inserted_:
			response = &indexedmapv1.EventsResponse{
				Event: indexedmapv1.Event{
					Event: &indexedmapv1.Event_Inserted_{
						Inserted: &indexedmapv1.Event_Inserted{
							Value: *newValue(&e.Inserted.Value),
						},
					},
				},
			}
		case *api.Event_Updated_:
			response = &indexedmapv1.EventsResponse{
				Event: indexedmapv1.Event{
					Event: &indexedmapv1.Event_Updated_{
						Updated: &indexedmapv1.Event_Updated{
							Value:     *newValue(&e.Updated.Value),
							PrevValue: *newValue(&e.Updated.PrevValue),
						},
					},
				},
			}
		case *api.Event_Removed_:
			response = &indexedmapv1.EventsResponse{
				Event: indexedmapv1.Event{
					Event: &indexedmapv1.Event_Removed_{
						Removed: &indexedmapv1.Event_Removed{
							Value:   *newValue(&e.Removed.Value),
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
}

func newEntry(entry *api.Entry) *indexedmapv1.Entry {
	if entry == nil {
		return nil
	}
	return &indexedmapv1.Entry{
		Key:   entry.Key,
		Index: entry.Index,
		Value: newValue(entry.Value),
	}
}

func newValue(value *api.Value) *indexedmapv1.VersionedValue {
	if value == nil {
		return nil
	}
	return &indexedmapv1.VersionedValue{
		Value:   value.Value,
		Version: value.Version,
	}
}

var _ indexedmapv1.IndexedMapServer = (*multiRaftIndexedMapServer)(nil)
