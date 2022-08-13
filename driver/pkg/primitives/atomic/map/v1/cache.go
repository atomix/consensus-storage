// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"container/list"
	"context"
	api "github.com/atomix/multi-raft-storage/api/atomix/multiraft/atomic/map/v1"
	atomicmapv1 "github.com/atomix/runtime/api/atomix/runtime/atomic/map/v1"
	"google.golang.org/grpc"
	"sync"
	"time"
)

func newCachingAtomicMapServer(m atomicmapv1.AtomicMapServer, config api.CacheConfig) atomicmapv1.AtomicMapServer {
	return &cachingAtomicMapServer{
		AtomicMapServer: m,
		config:          config,
		entries:         make(map[string]*list.Element),
		aged:            list.New(),
	}
}

type cachingAtomicMapServer struct {
	atomicmapv1.AtomicMapServer
	config  api.CacheConfig
	entries map[string]*list.Element
	aged    *list.List
	mu      sync.RWMutex
}

func (s *cachingAtomicMapServer) Create(ctx context.Context, request *atomicmapv1.CreateRequest) (*atomicmapv1.CreateResponse, error) {
	response, err := s.AtomicMapServer.Create(ctx, request)
	if err != nil {
		return nil, err
	}
	err = s.AtomicMapServer.Events(&atomicmapv1.EventsRequest{
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

func (s *cachingAtomicMapServer) Put(ctx context.Context, request *atomicmapv1.PutRequest) (*atomicmapv1.PutResponse, error) {
	response, err := s.AtomicMapServer.Put(ctx, request)
	if err != nil {
		return nil, err
	}
	s.update(&atomicmapv1.Entry{
		Key: request.Key,
		Value: &atomicmapv1.Value{
			Value:   request.Value,
			Version: response.NewVersion,
		},
	})
	return response, nil
}

func (s *cachingAtomicMapServer) Insert(ctx context.Context, request *atomicmapv1.InsertRequest) (*atomicmapv1.InsertResponse, error) {
	response, err := s.AtomicMapServer.Insert(ctx, request)
	if err != nil {
		return nil, err
	}
	s.update(&atomicmapv1.Entry{
		Key: request.Key,
		Value: &atomicmapv1.Value{
			Value:   request.Value,
			Version: response.NewVersion,
		},
	})
	return response, nil
}

func (s *cachingAtomicMapServer) Update(ctx context.Context, request *atomicmapv1.UpdateRequest) (*atomicmapv1.UpdateResponse, error) {
	response, err := s.AtomicMapServer.Update(ctx, request)
	if err != nil {
		return nil, err
	}
	s.update(&atomicmapv1.Entry{
		Key: request.Key,
		Value: &atomicmapv1.Value{
			Value:   request.Value,
			Version: response.NewVersion,
		},
	})
	return response, nil
}

func (s *cachingAtomicMapServer) Get(ctx context.Context, request *atomicmapv1.GetRequest) (*atomicmapv1.GetResponse, error) {
	s.mu.RLock()
	elem, ok := s.entries[request.Key]
	s.mu.RUnlock()
	if ok {
		return &atomicmapv1.GetResponse{
			Value: *elem.Value.(*cachedEntry).entry.Value,
		}, nil
	}
	return s.AtomicMapServer.Get(ctx, request)
}

func (s *cachingAtomicMapServer) Remove(ctx context.Context, request *atomicmapv1.RemoveRequest) (*atomicmapv1.RemoveResponse, error) {
	response, err := s.AtomicMapServer.Remove(ctx, request)
	if err != nil {
		return nil, err
	}
	s.delete(request.Key)
	return response, nil
}

func (s *cachingAtomicMapServer) Clear(ctx context.Context, request *atomicmapv1.ClearRequest) (*atomicmapv1.ClearResponse, error) {
	response, err := s.AtomicMapServer.Clear(ctx, request)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	s.entries = make(map[string]*list.Element)
	s.aged = list.New()
	s.mu.Unlock()
	return response, nil
}

func (s *cachingAtomicMapServer) update(update *atomicmapv1.Entry) {
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

func (s *cachingAtomicMapServer) delete(key string) {
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

func (s *cachingAtomicMapServer) evict() {
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

var _ atomicmapv1.AtomicMapServer = (*cachingAtomicMapServer)(nil)

func newCachingEventsServer(server *cachingAtomicMapServer) atomicmapv1.AtomicMap_EventsServer {
	return &cachingEventsServer{
		server: server,
	}
}

type cachingEventsServer struct {
	grpc.ServerStream
	server *cachingAtomicMapServer
}

func (s *cachingEventsServer) Send(response *atomicmapv1.EventsResponse) error {
	switch e := response.Event.Event.(type) {
	case *atomicmapv1.Event_Inserted_:
		s.server.update(&atomicmapv1.Entry{
			Key: response.Event.Key,
			Value: &atomicmapv1.Value{
				Value:   e.Inserted.Value.Value,
				Version: e.Inserted.Value.Version,
			},
		})
	case *atomicmapv1.Event_Updated_:
		s.server.update(&atomicmapv1.Entry{
			Key: response.Event.Key,
			Value: &atomicmapv1.Value{
				Value:   e.Updated.NewValue.Value,
				Version: e.Updated.NewValue.Version,
			},
		})
	case *atomicmapv1.Event_Removed_:
		s.server.delete(response.Event.Key)
	}
	return nil
}

func newCachedEntry(entry *atomicmapv1.Entry) *cachedEntry {
	return &cachedEntry{
		entry:     entry,
		timestamp: time.Now(),
	}
}

type cachedEntry struct {
	entry     *atomicmapv1.Entry
	timestamp time.Time
}
