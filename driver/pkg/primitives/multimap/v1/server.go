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
	"github.com/atomix/runtime/sdk/pkg/stringer"
	"google.golang.org/grpc"
	"io"
)

var log = logging.GetLogger()

const Service = "atomix.runtime.multimap.v1.MultiMap"

const truncLen = 200

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
		logging.Stringer("CreateRequest", stringer.Truncate(request, truncLen)))
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
			logging.Stringer("CreateRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &atomicmultimapv1.CreateResponse{}
	log.Debugw("Create",
		logging.Stringer("CreateRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("CreateResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *multiRaftMultiMapServer) Close(ctx context.Context, request *atomicmultimapv1.CloseRequest) (*atomicmultimapv1.CloseResponse, error) {
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
	response := &atomicmultimapv1.CloseResponse{}
	log.Debugw("Close",
		logging.Stringer("CloseRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("CloseResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *multiRaftMultiMapServer) Size(ctx context.Context, request *atomicmultimapv1.SizeRequest) (*atomicmultimapv1.SizeResponse, error) {
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
			return api.NewMultiMapClient(conn).Size(ctx, &api.SizeRequest{
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
	response := &atomicmultimapv1.SizeResponse{
		Size_: uint32(size),
	}
	log.Debugw("Size",
		logging.Stringer("SizeRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("SizeResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *multiRaftMultiMapServer) Put(ctx context.Context, request *atomicmultimapv1.PutRequest) (*atomicmultimapv1.PutResponse, error) {
	log.Debugw("Put",
		logging.Stringer("PutRequest", stringer.Truncate(request, truncLen)))
	partition := s.PartitionBy([]byte(request.Key))
	session, err := partition.GetSession(ctx)
	if err != nil {
		log.Warnw("Put",
			logging.Stringer("PutRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	primitive, err := session.GetPrimitive(request.ID.Name)
	if err != nil {
		log.Warnw("Put",
			logging.Stringer("PutRequest", stringer.Truncate(request, truncLen)),
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
			logging.Stringer("PutRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &atomicmultimapv1.PutResponse{}
	log.Debugw("Put",
		logging.Stringer("PutRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("PutResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *multiRaftMultiMapServer) PutAll(ctx context.Context, request *atomicmultimapv1.PutAllRequest) (*atomicmultimapv1.PutAllResponse, error) {
	log.Debugw("PutAll",
		logging.Stringer("PutAllRequest", stringer.Truncate(request, truncLen)))
	partition := s.PartitionBy([]byte(request.Key))
	session, err := partition.GetSession(ctx)
	if err != nil {
		log.Warnw("PutAll",
			logging.Stringer("PutAllRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	primitive, err := session.GetPrimitive(request.ID.Name)
	if err != nil {
		log.Warnw("PutAll",
			logging.Stringer("PutAllRequest", stringer.Truncate(request, truncLen)),
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
			logging.Stringer("PutAllRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &atomicmultimapv1.PutAllResponse{
		Updated: output.Updated,
	}
	log.Debugw("PutAll",
		logging.Stringer("PutAllRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("PutAllResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *multiRaftMultiMapServer) PutEntries(ctx context.Context, request *atomicmultimapv1.PutEntriesRequest) (*atomicmultimapv1.PutEntriesResponse, error) {
	log.Debugw("PutEntries",
		logging.Stringer("PutEntriesRequest", stringer.Truncate(request, truncLen)))
	entries := make(map[int][]api.Entry)
	for _, entry := range request.Entries {
		index := s.PartitionIndex([]byte(entry.Key))
		entries[index] = append(entries[index], api.Entry{
			Key:    entry.Key,
			Values: entry.Values,
		})
	}

	indexes := make([]int, 0, len(entries))
	for index := range entries {
		indexes = append(indexes, index)
	}

	partitions := s.Partitions()
	results, err := async.ExecuteAsync[bool](len(indexes), func(i int) (bool, error) {
		index := indexes[i]
		partition := partitions[index]

		session, err := partition.GetSession(ctx)
		if err != nil {
			log.Warnw("PutEntries",
				logging.Stringer("PutEntriesRequest", stringer.Truncate(request, truncLen)),
				logging.Error("Error", err))
			return false, err
		}
		primitive, err := session.GetPrimitive(request.ID.Name)
		if err != nil {
			log.Warnw("PutEntries",
				logging.Stringer("PutEntriesRequest", stringer.Truncate(request, truncLen)),
				logging.Error("Error", err))
			return false, err
		}
		command := client.Command[*api.PutEntriesResponse](primitive)
		input := &api.PutEntriesInput{
			Entries: entries[index],
		}
		output, err := command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (*api.PutEntriesResponse, error) {
			return api.NewMultiMapClient(conn).PutEntries(ctx, &api.PutEntriesRequest{
				Headers:         headers,
				PutEntriesInput: input,
			})
		})
		if err != nil {
			log.Warnw("PutEntries",
				logging.Stringer("PutEntriesRequest", stringer.Truncate(request, truncLen)),
				logging.Error("Error", err))
			return false, err
		}
		return output.Updated, nil
	})
	if err != nil {
		return nil, errors.ToProto(err)
	}
	response := &atomicmultimapv1.PutEntriesResponse{}
	for _, updated := range results {
		if updated {
			response.Updated = true
		}
	}
	log.Debugw("PutEntries",
		logging.Stringer("PutEntriesRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("PutEntriesResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *multiRaftMultiMapServer) Replace(ctx context.Context, request *atomicmultimapv1.ReplaceRequest) (*atomicmultimapv1.ReplaceResponse, error) {
	log.Debugw("Replace",
		logging.Stringer("ReplaceRequest", stringer.Truncate(request, truncLen)))
	partition := s.PartitionBy([]byte(request.Key))
	session, err := partition.GetSession(ctx)
	if err != nil {
		log.Warnw("Replace",
			logging.Stringer("ReplaceRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	primitive, err := session.GetPrimitive(request.ID.Name)
	if err != nil {
		log.Warnw("Replace",
			logging.Stringer("ReplaceRequest", stringer.Truncate(request, truncLen)),
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
			logging.Stringer("ReplaceRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &atomicmultimapv1.ReplaceResponse{
		PrevValues: output.PrevValues,
	}
	log.Debugw("Replace",
		logging.Stringer("ReplaceRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("ReplaceResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *multiRaftMultiMapServer) Contains(ctx context.Context, request *atomicmultimapv1.ContainsRequest) (*atomicmultimapv1.ContainsResponse, error) {
	log.Debugw("Contains",
		logging.Stringer("ContainsRequest", stringer.Truncate(request, truncLen)))
	partition := s.PartitionBy([]byte(request.Key))
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
			logging.Stringer("ContainsRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &atomicmultimapv1.ContainsResponse{
		Result: output.Result,
	}
	log.Debugw("Contains",
		logging.Stringer("ContainsRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("ContainsResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *multiRaftMultiMapServer) Get(ctx context.Context, request *atomicmultimapv1.GetRequest) (*atomicmultimapv1.GetResponse, error) {
	log.Debugw("Get",
		logging.Stringer("GetRequest", stringer.Truncate(request, truncLen)))
	partition := s.PartitionBy([]byte(request.Key))
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
		return api.NewMultiMapClient(conn).Get(ctx, &api.GetRequest{
			Headers: headers,
			GetInput: &api.GetInput{
				Key: request.Key,
			},
		})
	})
	if err != nil {
		log.Warnw("Get",
			logging.Stringer("GetRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &atomicmultimapv1.GetResponse{
		Values: output.Values,
	}
	log.Debugw("Get",
		logging.Stringer("GetRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("GetResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *multiRaftMultiMapServer) Remove(ctx context.Context, request *atomicmultimapv1.RemoveRequest) (*atomicmultimapv1.RemoveResponse, error) {
	log.Debugw("Remove",
		logging.Stringer("RemoveRequest", stringer.Truncate(request, truncLen)))
	partition := s.PartitionBy([]byte(request.Key))
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
			logging.Stringer("RemoveRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &atomicmultimapv1.RemoveResponse{
		Values: output.Values,
	}
	log.Debugw("Remove",
		logging.Stringer("RemoveRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("RemoveResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *multiRaftMultiMapServer) RemoveAll(ctx context.Context, request *atomicmultimapv1.RemoveAllRequest) (*atomicmultimapv1.RemoveAllResponse, error) {
	log.Debugw("RemoveAll",
		logging.Stringer("RemoveAllRequest", stringer.Truncate(request, truncLen)))
	partition := s.PartitionBy([]byte(request.Key))
	session, err := partition.GetSession(ctx)
	if err != nil {
		log.Warnw("RemoveAll",
			logging.Stringer("RemoveAllRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	primitive, err := session.GetPrimitive(request.ID.Name)
	if err != nil {
		log.Warnw("RemoveAll",
			logging.Stringer("RemoveAllRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	command := client.Command[*api.RemoveAllResponse](primitive)
	output, err := command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (*api.RemoveAllResponse, error) {
		input := &api.RemoveAllRequest{
			Headers: headers,
			RemoveAllInput: &api.RemoveAllInput{
				Key:    request.Key,
				Values: request.Values,
			},
		}
		return api.NewMultiMapClient(conn).RemoveAll(ctx, input)
	})
	if err != nil {
		log.Warnw("RemoveAll",
			logging.Stringer("RemoveAllRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &atomicmultimapv1.RemoveAllResponse{
		Updated: output.Updated,
	}
	log.Debugw("RemoveAll",
		logging.Stringer("RemoveAllRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("RemoveAllResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *multiRaftMultiMapServer) RemoveEntries(ctx context.Context, request *atomicmultimapv1.RemoveEntriesRequest) (*atomicmultimapv1.RemoveEntriesResponse, error) {
	log.Debugw("RemoveEntries",
		logging.Stringer("RemoveEntriesRequest", stringer.Truncate(request, truncLen)))
	entries := make(map[int][]api.Entry)
	for _, entry := range request.Entries {
		index := s.PartitionIndex([]byte(entry.Key))
		entries[index] = append(entries[index], api.Entry{
			Key:    entry.Key,
			Values: entry.Values,
		})
	}

	indexes := make([]int, 0, len(entries))
	for index := range entries {
		indexes = append(indexes, index)
	}

	partitions := s.Partitions()
	results, err := async.ExecuteAsync[bool](len(indexes), func(i int) (bool, error) {
		index := indexes[i]
		partition := partitions[index]

		session, err := partition.GetSession(ctx)
		if err != nil {
			log.Warnw("RemoveEntries",
				logging.Stringer("RemoveEntriesRequest", stringer.Truncate(request, truncLen)),
				logging.Error("Error", err))
			return false, err
		}
		primitive, err := session.GetPrimitive(request.ID.Name)
		if err != nil {
			log.Warnw("RemoveEntries",
				logging.Stringer("RemoveEntriesRequest", stringer.Truncate(request, truncLen)),
				logging.Error("Error", err))
			return false, err
		}
		command := client.Command[*api.RemoveEntriesResponse](primitive)
		input := &api.RemoveEntriesInput{
			Entries: entries[index],
		}
		output, err := command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (*api.RemoveEntriesResponse, error) {
			return api.NewMultiMapClient(conn).RemoveEntries(ctx, &api.RemoveEntriesRequest{
				Headers:            headers,
				RemoveEntriesInput: input,
			})
		})
		if err != nil {
			log.Warnw("RemoveEntries",
				logging.Stringer("RemoveEntriesRequest", stringer.Truncate(request, truncLen)),
				logging.Error("Error", err))
			return false, err
		}
		return output.Updated, nil
	})
	if err != nil {
		return nil, errors.ToProto(err)
	}
	response := &atomicmultimapv1.RemoveEntriesResponse{}
	for _, updated := range results {
		if updated {
			response.Updated = true
		}
	}
	log.Debugw("RemoveEntries",
		logging.Stringer("RemoveEntriesRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("RemoveEntriesResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *multiRaftMultiMapServer) Clear(ctx context.Context, request *atomicmultimapv1.ClearRequest) (*atomicmultimapv1.ClearResponse, error) {
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
			return api.NewMultiMapClient(conn).Clear(ctx, &api.ClearRequest{
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
	response := &atomicmultimapv1.ClearResponse{}
	log.Debugw("Clear",
		logging.Stringer("ClearRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("ClearResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *multiRaftMultiMapServer) Events(request *atomicmultimapv1.EventsRequest, server atomicmultimapv1.MultiMap_EventsServer) error {
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
			var response *atomicmultimapv1.EventsResponse
			switch e := output.Event.Event.(type) {
			case *api.Event_Added_:
				response = &atomicmultimapv1.EventsResponse{
					Event: atomicmultimapv1.Event{
						Key: output.Event.Key,
						Event: &atomicmultimapv1.Event_Added_{
							Added: &atomicmultimapv1.Event_Added{
								Value: e.Added.Value,
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
								Value: e.Removed.Value,
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

func (s *multiRaftMultiMapServer) Entries(request *atomicmultimapv1.EntriesRequest, server atomicmultimapv1.MultiMap_EntriesServer) error {
	log.Debugw("Entries",
		logging.Stringer("EntriesRequest", stringer.Truncate(request, truncLen)))
	partitions := s.Partitions()
	return async.IterAsync(len(partitions), func(i int) error {
		partition := partitions[i]
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
			return api.NewMultiMapClient(conn).Entries(server.Context(), &api.EntriesRequest{
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
			response := &atomicmultimapv1.EntriesResponse{
				Entry: atomicmultimapv1.Entry{
					Key:    output.Entry.Key,
					Values: output.Entry.Values,
				},
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
	})
}

var _ atomicmultimapv1.MultiMapServer = (*multiRaftMultiMapServer)(nil)
