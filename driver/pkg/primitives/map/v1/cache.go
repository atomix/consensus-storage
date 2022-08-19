// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"container/list"
	"context"
	api "github.com/atomix/multi-raft-storage/api/atomix/multiraft/map/v1"
	mapv1 "github.com/atomix/runtime/api/atomix/runtime/map/v1"
	"google.golang.org/grpc"
	"sync"
	"time"
)

func newCachingMapServer(m mapv1.MapServer, config api.CacheConfig) mapv1.MapServer {
	return &cachingMapServer{
		MapServer: m,
		config:    config,
		entries:   make(map[string]*list.Element),
		aged:      list.New(),
	}
}

type cachingMapServer struct {
	mapv1.MapServer
	config  api.CacheConfig
	entries map[string]*list.Element
	aged    *list.List
	mu      sync.RWMutex
}

func (s *cachingMapServer) Create(ctx context.Context, request *mapv1.CreateRequest) (*mapv1.CreateResponse, error) {
	response, err := s.MapServer.Create(ctx, request)
	if err != nil {
		return nil, err
	}
	err = s.MapServer.Events(&mapv1.EventsRequest{
		ID: request.ID,
	}, newCachingEventsServer(s))
	go func() {
		interval := s.config.EvictionInterval
		if interval == nil {
			i := time.Minute
			interval = &i
		}
		ticker := time.NewTicker(*interval)
		for range ticker.C {
			s.evict()
		}
	}()
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (s *cachingMapServer) Put(ctx context.Context, request *mapv1.PutRequest) (*mapv1.PutResponse, error) {
	response, err := s.MapServer.Put(ctx, request)
	if err != nil {
		return nil, err
	}
	s.update(&mapv1.Entry{
		Key: request.Key,
		Value: &mapv1.VersionedValue{
			Value:   request.Value,
			Version: response.Version,
		},
	})
	return response, nil
}

func (s *cachingMapServer) Insert(ctx context.Context, request *mapv1.InsertRequest) (*mapv1.InsertResponse, error) {
	response, err := s.MapServer.Insert(ctx, request)
	if err != nil {
		return nil, err
	}
	s.update(&mapv1.Entry{
		Key: request.Key,
		Value: &mapv1.VersionedValue{
			Value:   request.Value,
			Version: response.Version,
		},
	})
	return response, nil
}

func (s *cachingMapServer) Update(ctx context.Context, request *mapv1.UpdateRequest) (*mapv1.UpdateResponse, error) {
	response, err := s.MapServer.Update(ctx, request)
	if err != nil {
		return nil, err
	}
	s.update(&mapv1.Entry{
		Key: request.Key,
		Value: &mapv1.VersionedValue{
			Value:   request.Value,
			Version: response.Version,
		},
	})
	return response, nil
}

func (s *cachingMapServer) Get(ctx context.Context, request *mapv1.GetRequest) (*mapv1.GetResponse, error) {
	s.mu.RLock()
	elem, ok := s.entries[request.Key]
	s.mu.RUnlock()
	if ok {
		return &mapv1.GetResponse{
			Value: *elem.Value.(*cachedEntry).entry.Value,
		}, nil
	}
	return s.MapServer.Get(ctx, request)
}

func (s *cachingMapServer) Remove(ctx context.Context, request *mapv1.RemoveRequest) (*mapv1.RemoveResponse, error) {
	response, err := s.MapServer.Remove(ctx, request)
	if err != nil {
		return nil, err
	}
	s.delete(request.Key)
	return response, nil
}

func (s *cachingMapServer) Clear(ctx context.Context, request *mapv1.ClearRequest) (*mapv1.ClearResponse, error) {
	response, err := s.MapServer.Clear(ctx, request)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	s.entries = make(map[string]*list.Element)
	s.aged = list.New()
	s.mu.Unlock()
	return response, nil
}

func (s *cachingMapServer) update(update *mapv1.Entry) {
	s.mu.RLock()
	check, ok := s.entries[update.Key]
	s.mu.RUnlock()
	if ok && check.Value.(*cachedEntry).entry.Value.Version >= update.Value.Version {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	check, ok = s.entries[update.Key]
	if ok && check.Value.(*cachedEntry).entry.Value.Version >= update.Value.Version {
		return
	}

	entry := newCachedEntry(update)
	if elem, ok := s.entries[update.Key]; ok {
		s.aged.Remove(elem)
	}
	s.entries[update.Key] = s.aged.PushBack(entry)
}

func (s *cachingMapServer) delete(key string) {
	s.mu.RLock()
	_, ok := s.entries[key]
	s.mu.RUnlock()
	if !ok {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if elem, ok := s.entries[key]; ok {
		delete(s.entries, key)
		s.aged.Remove(elem)
	}
}

func (s *cachingMapServer) evict() {
	t := time.Now()
	evictionDuration := time.Hour
	if s.config.EvictAfter != nil {
		evictionDuration = *s.config.EvictAfter
	}

	s.mu.RLock()
	size := uint64(len(s.entries))
	entry := s.aged.Front()
	s.mu.RUnlock()
	if (entry == nil || t.Sub(entry.Value.(*cachedEntry).timestamp) < evictionDuration) && size <= s.config.Size_ {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	size = uint64(len(s.entries))
	entry = s.aged.Front()
	if (entry == nil || t.Sub(entry.Value.(*cachedEntry).timestamp) < evictionDuration) && size <= s.config.Size_ {
		return
	}

	for i := s.aged.Len(); i > int(s.config.Size_); i-- {
		if entry := s.aged.Front(); entry != nil {
			s.aged.Remove(entry)
		}
	}

	entry = s.aged.Front()
	for entry != nil {
		if t.Sub(entry.Value.(*cachedEntry).timestamp) < evictionDuration {
			break
		}
		s.aged.Remove(entry)
		entry = s.aged.Front()
	}
}

var _ mapv1.MapServer = (*cachingMapServer)(nil)

func newCachingEventsServer(server *cachingMapServer) mapv1.Map_EventsServer {
	return &cachingEventsServer{
		server: server,
	}
}

type cachingEventsServer struct {
	grpc.ServerStream
	server *cachingMapServer
}

func (s *cachingEventsServer) Send(response *mapv1.EventsResponse) error {
	switch e := response.Event.Event.(type) {
	case *mapv1.Event_Inserted_:
		s.server.update(&mapv1.Entry{
			Key: response.Event.Key,
			Value: &mapv1.VersionedValue{
				Value:   e.Inserted.Value.Value,
				Version: e.Inserted.Value.Version,
			},
		})
	case *mapv1.Event_Updated_:
		s.server.update(&mapv1.Entry{
			Key: response.Event.Key,
			Value: &mapv1.VersionedValue{
				Value:   e.Updated.Value.Value,
				Version: e.Updated.Value.Version,
			},
		})
	case *mapv1.Event_Removed_:
		s.server.delete(response.Event.Key)
	}
	return nil
}

func newCachedEntry(entry *mapv1.Entry) *cachedEntry {
	return &cachedEntry{
		entry:     entry,
		timestamp: time.Now(),
	}
}

type cachedEntry struct {
	entry     *mapv1.Entry
	timestamp time.Time
}
