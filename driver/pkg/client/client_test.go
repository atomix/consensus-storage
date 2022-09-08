// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/runtime/sdk/pkg/runtime"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"testing"
)

func TestClientPrimitiveCommandQuery(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	partitionServer := NewMockPartitionServer(ctrl)
	sessionServer := NewMockSessionServer(ctrl)
	testServer := NewMockTestServer(ctrl)

	network := runtime.NewLocalNetwork()
	lis, err := network.Listen("localhost:5678")

	server := grpc.NewServer()
	multiraftv1.RegisterPartitionServer(server, partitionServer)
	multiraftv1.RegisterSessionServer(server, sessionServer)
	RegisterTestServer(server, testServer)
	go func() {
		assert.NoError(t, server.Serve(lis))
	}()

	client := NewClient(network)
	err = client.Connect(context.TODO(), multiraftv1.DriverConfig{
		Partitions: []multiraftv1.PartitionConfig{
			{
				PartitionID: 1,
				Leader:      "localhost:5678",
			},
		},
	})
	assert.NoError(t, err)

	partition := client.Partitions()[0]
	partitionServer.EXPECT().OpenSession(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, request *multiraftv1.OpenSessionRequest) (*multiraftv1.OpenSessionResponse, error) {
			return &multiraftv1.OpenSessionResponse{
				Headers: &multiraftv1.PartitionResponseHeaders{
					Index: 1,
				},
				OpenSessionOutput: &multiraftv1.OpenSessionOutput{
					SessionID: 1,
				},
			}, nil
		})
	session, err := partition.GetSession(context.TODO())
	assert.NoError(t, err)

	sessionServer.EXPECT().CreatePrimitive(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, request *multiraftv1.CreatePrimitiveRequest) (*multiraftv1.CreatePrimitiveResponse, error) {
			assert.Equal(t, multiraftv1.PartitionID(1), request.Headers.PartitionID)
			assert.Equal(t, multiraftv1.SessionID(1), request.Headers.SessionID)
			assert.Equal(t, multiraftv1.SequenceNum(1), request.Headers.SequenceNum)
			return &multiraftv1.CreatePrimitiveResponse{
				Headers: &multiraftv1.CommandResponseHeaders{
					OperationResponseHeaders: multiraftv1.OperationResponseHeaders{
						PrimitiveResponseHeaders: multiraftv1.PrimitiveResponseHeaders{
							SessionResponseHeaders: multiraftv1.SessionResponseHeaders{
								PartitionResponseHeaders: multiraftv1.PartitionResponseHeaders{
									Index: 2,
								},
							},
						},
						Status: multiraftv1.OperationResponseHeaders_OK,
					},
					OutputSequenceNum: 1,
				},
				CreatePrimitiveOutput: &multiraftv1.CreatePrimitiveOutput{
					PrimitiveID: 2,
				},
			}, nil
		})
	err = session.CreatePrimitive(context.TODO(), "name", "service")
	assert.NoError(t, err)

	primitive, err := session.GetPrimitive("name")
	assert.NoError(t, err)

	command := Command[*TestCommandResponse](primitive)
	testServer.EXPECT().TestCommand(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, request *TestCommandRequest) (*TestCommandResponse, error) {
			assert.Equal(t, multiraftv1.PartitionID(1), request.Headers.PartitionID)
			assert.Equal(t, multiraftv1.SessionID(1), request.Headers.SessionID)
			assert.Equal(t, multiraftv1.SequenceNum(2), request.Headers.SequenceNum)
			assert.Equal(t, multiraftv1.PrimitiveID(2), request.Headers.PrimitiveID)
			return &TestCommandResponse{
				Headers: &multiraftv1.CommandResponseHeaders{
					OperationResponseHeaders: multiraftv1.OperationResponseHeaders{
						PrimitiveResponseHeaders: multiraftv1.PrimitiveResponseHeaders{
							SessionResponseHeaders: multiraftv1.SessionResponseHeaders{
								PartitionResponseHeaders: multiraftv1.PartitionResponseHeaders{
									Index: 3,
								},
							},
						},
						Status: multiraftv1.OperationResponseHeaders_OK,
					},
					OutputSequenceNum: 1,
				},
			}, nil
		})
	commandResponse, err := command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (*TestCommandResponse, error) {
		return NewTestClient(conn).TestCommand(context.TODO(), &TestCommandRequest{
			Headers: headers,
		})
	})
	assert.NoError(t, err)
	assert.NotNil(t, commandResponse)

	query := Query[*TestQueryResponse](primitive)
	testServer.EXPECT().TestQuery(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, request *TestQueryRequest) (*TestQueryResponse, error) {
			assert.Equal(t, multiraftv1.PartitionID(1), request.Headers.PartitionID)
			assert.Equal(t, multiraftv1.SessionID(1), request.Headers.SessionID)
			assert.Equal(t, multiraftv1.SequenceNum(3), request.Headers.SequenceNum)
			assert.Equal(t, multiraftv1.PrimitiveID(2), request.Headers.PrimitiveID)
			return &TestQueryResponse{
				Headers: &multiraftv1.QueryResponseHeaders{
					OperationResponseHeaders: multiraftv1.OperationResponseHeaders{
						PrimitiveResponseHeaders: multiraftv1.PrimitiveResponseHeaders{
							SessionResponseHeaders: multiraftv1.SessionResponseHeaders{
								PartitionResponseHeaders: multiraftv1.PartitionResponseHeaders{
									Index: 3,
								},
							},
						},
						Status: multiraftv1.OperationResponseHeaders_OK,
					},
				},
			}, nil
		})
	queryResponse, err := query.Run(func(conn *grpc.ClientConn, headers *multiraftv1.QueryRequestHeaders) (*TestQueryResponse, error) {
		return NewTestClient(conn).TestQuery(context.TODO(), &TestQueryRequest{
			Headers: headers,
		})
	})
	assert.NoError(t, err)
	assert.NotNil(t, queryResponse)

	sessionServer.EXPECT().ClosePrimitive(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, request *multiraftv1.ClosePrimitiveRequest) (*multiraftv1.ClosePrimitiveResponse, error) {
			assert.Equal(t, multiraftv1.PartitionID(1), request.Headers.PartitionID)
			assert.Equal(t, multiraftv1.SessionID(1), request.Headers.SessionID)
			assert.Equal(t, multiraftv1.PrimitiveID(2), request.PrimitiveID)
			assert.Equal(t, multiraftv1.SequenceNum(4), request.Headers.SequenceNum)
			return &multiraftv1.ClosePrimitiveResponse{
				Headers: &multiraftv1.CommandResponseHeaders{
					OperationResponseHeaders: multiraftv1.OperationResponseHeaders{
						PrimitiveResponseHeaders: multiraftv1.PrimitiveResponseHeaders{
							SessionResponseHeaders: multiraftv1.SessionResponseHeaders{
								PartitionResponseHeaders: multiraftv1.PartitionResponseHeaders{
									Index: 4,
								},
							},
						},
						Status: multiraftv1.OperationResponseHeaders_OK,
					},
					OutputSequenceNum: 1,
				},
			}, nil
		})
	err = session.ClosePrimitive(context.TODO(), "name")
	assert.NoError(t, err)
}
