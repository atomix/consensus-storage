// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"fmt"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/runtime/sdk/pkg/grpc/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/serviceconfig"
)

const resolverName = "raft"

func newResolver(partition *PartitionClient) resolver.Builder {
	return &ResolverBuilder{
		partition: partition,
	}
}

type ResolverBuilder struct {
	partition *PartitionClient
}

func (b *ResolverBuilder) Scheme() string {
	return resolverName
}

func (b *ResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	var dialOpts []grpc.DialOption
	if opts.DialCreds != nil {
		dialOpts = append(
			dialOpts,
			grpc.WithTransportCredentials(opts.DialCreds),
		)
	} else {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	dialOpts = append(dialOpts, grpc.WithUnaryInterceptor(retry.RetryingUnaryClientInterceptor(retry.WithRetryOn(codes.Unavailable, codes.Unknown))))
	dialOpts = append(dialOpts, grpc.WithStreamInterceptor(retry.RetryingStreamClientInterceptor(retry.WithRetryOn(codes.Unavailable, codes.Unknown))))
	dialOpts = append(dialOpts, grpc.WithContextDialer(opts.Dialer))

	resolverConn, err := grpc.Dial(target.Endpoint, dialOpts...)
	if err != nil {
		return nil, err
	}

	serviceConfig := cc.ParseServiceConfig(
		fmt.Sprintf(`{"loadBalancingConfig":[{"%s":{}}]}`, resolverName),
	)

	resolver := &Resolver{
		partitionID:   b.partition.ID(),
		clientConn:    cc,
		resolverConn:  resolverConn,
		serviceConfig: serviceConfig,
	}
	err = resolver.start()
	if err != nil {
		return nil, err
	}
	return resolver, nil
}

var _ resolver.Builder = (*ResolverBuilder)(nil)

type Resolver struct {
	partitionID   multiraftv1.PartitionID
	clientConn    resolver.ClientConn
	resolverConn  *grpc.ClientConn
	serviceConfig *serviceconfig.ParseResult
}

func (r *Resolver) start() error {
	client := multiraftv1.NewPartitionClient(r.resolverConn)
	request := &multiraftv1.WatchPartitionRequest{
		PartitionID: r.partitionID,
	}
	stream, err := client.Watch(context.Background(), request)
	if err != nil {
		return err
	}
	response, err := stream.Recv()
	if err != nil {
		return err
	}
	switch e := response.Event.(type) {
	case *multiraftv1.PartitionEvent_ServiceConfigChanged:
		r.updateState(e.ServiceConfigChanged.Config)
	}
	go func() {
		for {
			response, err := stream.Recv()
			if err != nil {
				return
			}
			switch e := response.Event.(type) {
			case *multiraftv1.PartitionEvent_ServiceConfigChanged:
				r.updateState(e.ServiceConfigChanged.Config)
			}
		}
	}()
	return nil
}

func (r *Resolver) updateState(config multiraftv1.ServiceConfig) {
	log.Debugf("Updating connections for partition config %v", config)

	var addrs []resolver.Address
	addrs = append(addrs, resolver.Address{
		Addr: config.Leader,
		Type: resolver.Backend,
		Attributes: attributes.New(
			"is_leader",
			true,
		),
	})

	for _, server := range config.Followers {
		addrs = append(addrs, resolver.Address{
			Addr: server,
			Type: resolver.Backend,
			Attributes: attributes.New(
				"is_leader",
				false,
			),
		})
	}

	r.clientConn.UpdateState(resolver.State{
		Addresses:     addrs,
		ServiceConfig: r.serviceConfig,
	})
}

func (r *Resolver) ResolveNow(resolver.ResolveNowOptions) {}

func (r *Resolver) Close() {
	if err := r.resolverConn.Close(); err != nil {
		log.Error("failed to close conn", err)
	}
}

var _ resolver.Resolver = (*Resolver)(nil)
