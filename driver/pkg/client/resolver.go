// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"fmt"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/serviceconfig"
)

const resolverName = "rsm"

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
	serviceConfig := cc.ParseServiceConfig(
		fmt.Sprintf(`{"loadBalancingConfig":[{"%s":{}}]}`, resolverName),
	)

	resolver := &Resolver{
		partition:     b.partition,
		clientConn:    cc,
		serviceConfig: serviceConfig,
	}
	err := resolver.start()
	if err != nil {
		return nil, err
	}
	return resolver, nil
}

var _ resolver.Builder = (*ResolverBuilder)(nil)

type Resolver struct {
	partition     *PartitionClient
	clientConn    resolver.ClientConn
	serviceConfig *serviceconfig.ParseResult
}

func (r *Resolver) start() error {
	ch := make(chan PartitionState)
	if err := r.partition.watch(context.Background(), ch); err != nil {
		return err
	}

	state := <-ch
	r.updateState(state)

	go func() {
		for state := range ch {
			r.updateState(state)
		}
	}()
	return nil
}

func (r *Resolver) updateState(state PartitionState) {
	log.Debugf("Updating connections for partition state %+v", state)

	var addrs []resolver.Address
	addrs = append(addrs, resolver.Address{
		Addr: state.Leader,
		Type: resolver.Backend,
		Attributes: attributes.New(
			"is_leader",
			true,
		),
	})

	for _, server := range state.Followers {
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

}

var _ resolver.Resolver = (*Resolver)(nil)
