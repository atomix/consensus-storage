// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"encoding/binary"
	"encoding/json"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/atomix/runtime/sdk/pkg/runtime"
	"github.com/bits-and-blooms/bloom/v3"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"io"
	"testing"
	"time"
)

func TestPrimitiveCreateClose(t *testing.T) {
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

	partitionServer.EXPECT().CloseSession(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, request *multiraftv1.CloseSessionRequest) (*multiraftv1.CloseSessionResponse, error) {
			return &multiraftv1.CloseSessionResponse{
				Headers: &multiraftv1.PartitionResponseHeaders{
					Index: 5,
				},
			}, nil
		})
	err = client.Close(context.TODO())
	assert.NoError(t, err)
}

func TestUnaryCommand(t *testing.T) {
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
	timeout := 10 * time.Second
	err = client.Connect(context.TODO(), multiraftv1.DriverConfig{
		Partitions: []multiraftv1.PartitionConfig{
			{
				PartitionID: 1,
				Leader:      "localhost:5678",
			},
		},
		SessionTimeout: &timeout,
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
		Return(nil, errors.ToProto(errors.NewUnavailable("unavailable")))
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

	ch := make(chan struct{})
	partitionServer.EXPECT().KeepAlive(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, request *multiraftv1.KeepAliveRequest) (*multiraftv1.KeepAliveResponse, error) {
			assert.Equal(t, multiraftv1.PartitionID(1), request.Headers.PartitionID)
			assert.Equal(t, multiraftv1.SessionID(1), request.SessionID)
			assert.Equal(t, multiraftv1.SequenceNum(2), request.LastInputSequenceNum)
			inputFilter := &bloom.BloomFilter{}
			assert.NoError(t, json.Unmarshal(request.InputFilter, inputFilter))
			sequenceNumBytes := make([]byte, 8)
			binary.BigEndian.PutUint64(sequenceNumBytes, 1)
			assert.False(t, inputFilter.Test(sequenceNumBytes))
			sequenceNumBytes = make([]byte, 8)
			binary.BigEndian.PutUint64(sequenceNumBytes, 2)
			assert.False(t, inputFilter.Test(sequenceNumBytes))
			close(ch)
			return &multiraftv1.KeepAliveResponse{
				Headers: &multiraftv1.PartitionResponseHeaders{
					Index: 4,
				},
			}, nil
		})
	select {
	case <-ch:
	case <-time.After(time.Minute):
		t.FailNow()
	}

	partitionServer.EXPECT().CloseSession(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, request *multiraftv1.CloseSessionRequest) (*multiraftv1.CloseSessionResponse, error) {
			return &multiraftv1.CloseSessionResponse{
				Headers: &multiraftv1.PartitionResponseHeaders{
					Index: 5,
				},
			}, nil
		})
	err = client.Close(context.TODO())
	assert.NoError(t, err)
}

func TestStreamCommand(t *testing.T) {
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
	timeout := 10 * time.Second
	err = client.Connect(context.TODO(), multiraftv1.DriverConfig{
		Partitions: []multiraftv1.PartitionConfig{
			{
				PartitionID: 1,
				Leader:      "localhost:5678",
			},
		},
		SessionTimeout: &timeout,
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

	command := StreamCommand[*TestCommandResponse](primitive)
	sendResponseCh := make(chan struct{})
	testServer.EXPECT().TestStreamCommand(gomock.Any(), gomock.Any()).
		Return(errors.ToProto(errors.NewUnavailable("unavailable")))
	testServer.EXPECT().TestStreamCommand(gomock.Any(), gomock.Any()).
		DoAndReturn(func(request *TestCommandRequest, stream Test_TestStreamCommandServer) error {
			assert.Equal(t, multiraftv1.PartitionID(1), request.Headers.PartitionID)
			assert.Equal(t, multiraftv1.SessionID(1), request.Headers.SessionID)
			assert.Equal(t, multiraftv1.SequenceNum(2), request.Headers.SequenceNum)
			assert.Equal(t, multiraftv1.PrimitiveID(2), request.Headers.PrimitiveID)
			var outputSequenceNum multiraftv1.SequenceNum
			for range sendResponseCh {
				outputSequenceNum++
				assert.Nil(t, stream.Send(&TestCommandResponse{
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
						OutputSequenceNum: outputSequenceNum,
					},
				}))
			}
			return nil
		})
	stream, err := command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (CommandStream[*TestCommandResponse], error) {
		return NewTestClient(conn).TestStreamCommand(context.TODO(), &TestCommandRequest{
			Headers: headers,
		})
	})
	assert.NoError(t, err)

	go func() {
		sendResponseCh <- struct{}{}
	}()
	response, err := stream.Recv()
	assert.NoError(t, err)
	assert.NotNil(t, response)
	go func() {
		sendResponseCh <- struct{}{}
	}()
	response, err = stream.Recv()
	assert.NoError(t, err)
	assert.NotNil(t, response)

	keepAliveDone := make(chan struct{})
	partitionServer.EXPECT().KeepAlive(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, request *multiraftv1.KeepAliveRequest) (*multiraftv1.KeepAliveResponse, error) {
			assert.Equal(t, multiraftv1.PartitionID(1), request.Headers.PartitionID)
			assert.Equal(t, multiraftv1.SessionID(1), request.SessionID)
			assert.Equal(t, multiraftv1.SequenceNum(2), request.LastInputSequenceNum)
			inputFilter := &bloom.BloomFilter{}
			assert.NoError(t, json.Unmarshal(request.InputFilter, inputFilter))
			sequenceNumBytes := make([]byte, 8)
			binary.BigEndian.PutUint64(sequenceNumBytes, 1)
			assert.False(t, inputFilter.Test(sequenceNumBytes))
			sequenceNumBytes = make([]byte, 8)
			binary.BigEndian.PutUint64(sequenceNumBytes, 2)
			assert.True(t, inputFilter.Test(sequenceNumBytes))
			assert.Equal(t, multiraftv1.SequenceNum(2), request.LastOutputSequenceNums[2])
			close(keepAliveDone)
			return &multiraftv1.KeepAliveResponse{
				Headers: &multiraftv1.PartitionResponseHeaders{
					Index: 4,
				},
			}, nil
		})
	select {
	case <-keepAliveDone:
	case <-time.After(time.Minute):
		t.FailNow()
	}

	go func() {
		sendResponseCh <- struct{}{}
	}()
	response, err = stream.Recv()
	assert.NoError(t, err)
	assert.NotNil(t, response)

	close(sendResponseCh)

	keepAliveDone = make(chan struct{})
	partitionServer.EXPECT().KeepAlive(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, request *multiraftv1.KeepAliveRequest) (*multiraftv1.KeepAliveResponse, error) {
			assert.Equal(t, multiraftv1.PartitionID(1), request.Headers.PartitionID)
			assert.Equal(t, multiraftv1.SessionID(1), request.SessionID)
			assert.Equal(t, multiraftv1.SequenceNum(2), request.LastInputSequenceNum)
			inputFilter := &bloom.BloomFilter{}
			assert.NoError(t, json.Unmarshal(request.InputFilter, inputFilter))
			sequenceNumBytes := make([]byte, 8)
			binary.BigEndian.PutUint64(sequenceNumBytes, 1)
			assert.False(t, inputFilter.Test(sequenceNumBytes))
			sequenceNumBytes = make([]byte, 8)
			binary.BigEndian.PutUint64(sequenceNumBytes, 2)
			assert.True(t, inputFilter.Test(sequenceNumBytes))
			assert.Equal(t, multiraftv1.SequenceNum(3), request.LastOutputSequenceNums[2])
			close(keepAliveDone)
			return &multiraftv1.KeepAliveResponse{
				Headers: &multiraftv1.PartitionResponseHeaders{
					Index: 5,
				},
			}, nil
		})
	select {
	case <-keepAliveDone:
	case <-time.After(time.Minute):
		t.FailNow()
	}

	response, err = stream.Recv()
	assert.Equal(t, io.EOF, err)
	assert.Nil(t, response)

	keepAliveDone = make(chan struct{})
	partitionServer.EXPECT().KeepAlive(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, request *multiraftv1.KeepAliveRequest) (*multiraftv1.KeepAliveResponse, error) {
			assert.Equal(t, multiraftv1.PartitionID(1), request.Headers.PartitionID)
			assert.Equal(t, multiraftv1.SessionID(1), request.SessionID)
			assert.Equal(t, multiraftv1.SequenceNum(2), request.LastInputSequenceNum)
			inputFilter := &bloom.BloomFilter{}
			assert.NoError(t, json.Unmarshal(request.InputFilter, inputFilter))
			sequenceNumBytes := make([]byte, 8)
			binary.BigEndian.PutUint64(sequenceNumBytes, 1)
			assert.False(t, inputFilter.Test(sequenceNumBytes))
			sequenceNumBytes = make([]byte, 8)
			binary.BigEndian.PutUint64(sequenceNumBytes, 2)
			assert.False(t, inputFilter.Test(sequenceNumBytes))
			_, ok := request.LastOutputSequenceNums[2]
			assert.False(t, ok)
			close(keepAliveDone)
			return &multiraftv1.KeepAliveResponse{
				Headers: &multiraftv1.PartitionResponseHeaders{
					Index: 6,
				},
			}, nil
		})
	select {
	case <-keepAliveDone:
	case <-time.After(time.Minute):
		t.FailNow()
	}

	partitionServer.EXPECT().CloseSession(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, request *multiraftv1.CloseSessionRequest) (*multiraftv1.CloseSessionResponse, error) {
			return &multiraftv1.CloseSessionResponse{
				Headers: &multiraftv1.PartitionResponseHeaders{
					Index: 7,
				},
			}, nil
		})
	err = client.Close(context.TODO())
	assert.NoError(t, err)
}

func TestStreamCommandCancel(t *testing.T) {
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
	timeout := 10 * time.Second
	err = client.Connect(context.TODO(), multiraftv1.DriverConfig{
		Partitions: []multiraftv1.PartitionConfig{
			{
				PartitionID: 1,
				Leader:      "localhost:5678",
			},
		},
		SessionTimeout: &timeout,
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

	command := StreamCommand[*TestCommandResponse](primitive)
	testServer.EXPECT().TestStreamCommand(gomock.Any(), gomock.Any()).
		Return(errors.ToProto(errors.NewUnavailable("unavailable")))
	testServer.EXPECT().TestStreamCommand(gomock.Any(), gomock.Any()).
		DoAndReturn(func(request *TestCommandRequest, stream Test_TestStreamCommandServer) error {
			assert.Equal(t, multiraftv1.PartitionID(1), request.Headers.PartitionID)
			assert.Equal(t, multiraftv1.SessionID(1), request.Headers.SessionID)
			assert.Equal(t, multiraftv1.SequenceNum(2), request.Headers.SequenceNum)
			assert.Equal(t, multiraftv1.PrimitiveID(2), request.Headers.PrimitiveID)
			assert.Nil(t, stream.Send(&TestCommandResponse{
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
			}))
			assert.Nil(t, stream.Send(&TestCommandResponse{
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
					OutputSequenceNum: 2,
				},
			}))
			<-stream.Context().Done()
			return stream.Context().Err()
		})

	ctx, cancel := context.WithCancel(context.Background())
	stream, err := command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (CommandStream[*TestCommandResponse], error) {
		return NewTestClient(conn).TestStreamCommand(ctx, &TestCommandRequest{
			Headers: headers,
		})
	})
	assert.NoError(t, err)

	response, err := stream.Recv()
	assert.NoError(t, err)
	assert.NotNil(t, response)
	response, err = stream.Recv()
	assert.NoError(t, err)
	assert.NotNil(t, response)

	keepAliveDone := make(chan struct{})
	partitionServer.EXPECT().KeepAlive(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, request *multiraftv1.KeepAliveRequest) (*multiraftv1.KeepAliveResponse, error) {
			assert.Equal(t, multiraftv1.PartitionID(1), request.Headers.PartitionID)
			assert.Equal(t, multiraftv1.SessionID(1), request.SessionID)
			assert.Equal(t, multiraftv1.SequenceNum(2), request.LastInputSequenceNum)
			inputFilter := &bloom.BloomFilter{}
			assert.NoError(t, json.Unmarshal(request.InputFilter, inputFilter))
			sequenceNumBytes := make([]byte, 8)
			binary.BigEndian.PutUint64(sequenceNumBytes, 1)
			assert.False(t, inputFilter.Test(sequenceNumBytes))
			sequenceNumBytes = make([]byte, 8)
			binary.BigEndian.PutUint64(sequenceNumBytes, 2)
			assert.True(t, inputFilter.Test(sequenceNumBytes))
			assert.Equal(t, multiraftv1.SequenceNum(2), request.LastOutputSequenceNums[2])
			close(keepAliveDone)
			return &multiraftv1.KeepAliveResponse{
				Headers: &multiraftv1.PartitionResponseHeaders{
					Index: 4,
				},
			}, nil
		})
	select {
	case <-keepAliveDone:
	case <-time.After(time.Minute):
		t.FailNow()
	}

	cancel()

	keepAliveDone = make(chan struct{})
	partitionServer.EXPECT().KeepAlive(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, request *multiraftv1.KeepAliveRequest) (*multiraftv1.KeepAliveResponse, error) {
			assert.Equal(t, multiraftv1.PartitionID(1), request.Headers.PartitionID)
			assert.Equal(t, multiraftv1.SessionID(1), request.SessionID)
			assert.Equal(t, multiraftv1.SequenceNum(2), request.LastInputSequenceNum)
			inputFilter := &bloom.BloomFilter{}
			assert.NoError(t, json.Unmarshal(request.InputFilter, inputFilter))
			sequenceNumBytes := make([]byte, 8)
			binary.BigEndian.PutUint64(sequenceNumBytes, 1)
			assert.False(t, inputFilter.Test(sequenceNumBytes))
			sequenceNumBytes = make([]byte, 8)
			binary.BigEndian.PutUint64(sequenceNumBytes, 2)
			assert.True(t, inputFilter.Test(sequenceNumBytes))
			close(keepAliveDone)
			return &multiraftv1.KeepAliveResponse{
				Headers: &multiraftv1.PartitionResponseHeaders{
					Index: 5,
				},
			}, nil
		})
	select {
	case <-keepAliveDone:
	case <-time.After(time.Minute):
		t.FailNow()
	}

	_, err = stream.Recv()
	assert.Error(t, err)

	keepAliveDone = make(chan struct{})
	partitionServer.EXPECT().KeepAlive(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, request *multiraftv1.KeepAliveRequest) (*multiraftv1.KeepAliveResponse, error) {
			assert.Equal(t, multiraftv1.PartitionID(1), request.Headers.PartitionID)
			assert.Equal(t, multiraftv1.SessionID(1), request.SessionID)
			assert.Equal(t, multiraftv1.SequenceNum(2), request.LastInputSequenceNum)
			inputFilter := &bloom.BloomFilter{}
			assert.NoError(t, json.Unmarshal(request.InputFilter, inputFilter))
			sequenceNumBytes := make([]byte, 8)
			binary.BigEndian.PutUint64(sequenceNumBytes, 1)
			assert.False(t, inputFilter.Test(sequenceNumBytes))
			sequenceNumBytes = make([]byte, 8)
			binary.BigEndian.PutUint64(sequenceNumBytes, 2)
			assert.False(t, inputFilter.Test(sequenceNumBytes))
			_, ok := request.LastOutputSequenceNums[2]
			assert.False(t, ok)
			close(keepAliveDone)
			return &multiraftv1.KeepAliveResponse{
				Headers: &multiraftv1.PartitionResponseHeaders{
					Index: 5,
				},
			}, nil
		})
	select {
	case <-keepAliveDone:
	case <-time.After(time.Minute):
		t.FailNow()
	}

	partitionServer.EXPECT().CloseSession(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, request *multiraftv1.CloseSessionRequest) (*multiraftv1.CloseSessionResponse, error) {
			return &multiraftv1.CloseSessionResponse{
				Headers: &multiraftv1.PartitionResponseHeaders{
					Index: 6,
				},
			}, nil
		})
	err = client.Close(context.TODO())
	assert.NoError(t, err)
}

func TestUnaryQuery(t *testing.T) {
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
	timeout := 10 * time.Second
	err = client.Connect(context.TODO(), multiraftv1.DriverConfig{
		Partitions: []multiraftv1.PartitionConfig{
			{
				PartitionID: 1,
				Leader:      "localhost:5678",
			},
		},
		SessionTimeout: &timeout,
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

	query := Query[*TestQueryResponse](primitive)
	testServer.EXPECT().TestQuery(gomock.Any(), gomock.Any()).
		Return(nil, errors.ToProto(errors.NewUnavailable("unavailable")))
	testServer.EXPECT().TestQuery(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, request *TestQueryRequest) (*TestQueryResponse, error) {
			assert.Equal(t, multiraftv1.PartitionID(1), request.Headers.PartitionID)
			assert.Equal(t, multiraftv1.SessionID(1), request.Headers.SessionID)
			assert.Equal(t, multiraftv1.SequenceNum(2), request.Headers.SequenceNum)
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
	response, err := query.Run(func(conn *grpc.ClientConn, headers *multiraftv1.QueryRequestHeaders) (*TestQueryResponse, error) {
		return NewTestClient(conn).TestQuery(context.TODO(), &TestQueryRequest{
			Headers: headers,
		})
	})
	assert.NoError(t, err)
	assert.NotNil(t, response)

	ch := make(chan struct{})
	partitionServer.EXPECT().KeepAlive(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, request *multiraftv1.KeepAliveRequest) (*multiraftv1.KeepAliveResponse, error) {
			assert.Equal(t, multiraftv1.PartitionID(1), request.Headers.PartitionID)
			assert.Equal(t, multiraftv1.SessionID(1), request.SessionID)
			assert.Equal(t, multiraftv1.SequenceNum(2), request.LastInputSequenceNum)
			inputFilter := &bloom.BloomFilter{}
			assert.NoError(t, json.Unmarshal(request.InputFilter, inputFilter))
			sequenceNumBytes := make([]byte, 8)
			binary.BigEndian.PutUint64(sequenceNumBytes, 1)
			assert.False(t, inputFilter.Test(sequenceNumBytes))
			sequenceNumBytes = make([]byte, 8)
			binary.BigEndian.PutUint64(sequenceNumBytes, 2)
			assert.False(t, inputFilter.Test(sequenceNumBytes))
			close(ch)
			return &multiraftv1.KeepAliveResponse{
				Headers: &multiraftv1.PartitionResponseHeaders{
					Index: 4,
				},
			}, nil
		})
	select {
	case <-ch:
	case <-time.After(time.Minute):
		t.FailNow()
	}

	partitionServer.EXPECT().CloseSession(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, request *multiraftv1.CloseSessionRequest) (*multiraftv1.CloseSessionResponse, error) {
			return &multiraftv1.CloseSessionResponse{
				Headers: &multiraftv1.PartitionResponseHeaders{
					Index: 5,
				},
			}, nil
		})
	err = client.Close(context.TODO())
	assert.NoError(t, err)
}

func TestStreamQuery(t *testing.T) {
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
	timeout := 10 * time.Second
	err = client.Connect(context.TODO(), multiraftv1.DriverConfig{
		Partitions: []multiraftv1.PartitionConfig{
			{
				PartitionID: 1,
				Leader:      "localhost:5678",
			},
		},
		SessionTimeout: &timeout,
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

	query := StreamQuery[*TestQueryResponse](primitive)
	sendResponseCh := make(chan struct{})
	testServer.EXPECT().TestStreamQuery(gomock.Any(), gomock.Any()).
		Return(errors.ToProto(errors.NewUnavailable("unavailable")))
	testServer.EXPECT().TestStreamQuery(gomock.Any(), gomock.Any()).
		DoAndReturn(func(request *TestQueryRequest, stream Test_TestStreamQueryServer) error {
			assert.Equal(t, multiraftv1.PartitionID(1), request.Headers.PartitionID)
			assert.Equal(t, multiraftv1.SessionID(1), request.Headers.SessionID)
			assert.Equal(t, multiraftv1.SequenceNum(2), request.Headers.SequenceNum)
			assert.Equal(t, multiraftv1.PrimitiveID(2), request.Headers.PrimitiveID)
			for range sendResponseCh {
				assert.Nil(t, stream.Send(&TestQueryResponse{
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
				}))
			}
			return nil
		})
	stream, err := query.Run(func(conn *grpc.ClientConn, headers *multiraftv1.QueryRequestHeaders) (QueryStream[*TestQueryResponse], error) {
		return NewTestClient(conn).TestStreamQuery(context.TODO(), &TestQueryRequest{
			Headers: headers,
		})
	})
	assert.NoError(t, err)

	go func() {
		sendResponseCh <- struct{}{}
	}()
	response, err := stream.Recv()
	assert.NoError(t, err)
	assert.NotNil(t, response)
	go func() {
		sendResponseCh <- struct{}{}
	}()
	response, err = stream.Recv()
	assert.NoError(t, err)
	assert.NotNil(t, response)

	keepAliveDone := make(chan struct{})
	partitionServer.EXPECT().KeepAlive(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, request *multiraftv1.KeepAliveRequest) (*multiraftv1.KeepAliveResponse, error) {
			assert.Equal(t, multiraftv1.PartitionID(1), request.Headers.PartitionID)
			assert.Equal(t, multiraftv1.SessionID(1), request.SessionID)
			assert.Equal(t, multiraftv1.SequenceNum(2), request.LastInputSequenceNum)
			inputFilter := &bloom.BloomFilter{}
			assert.NoError(t, json.Unmarshal(request.InputFilter, inputFilter))
			sequenceNumBytes := make([]byte, 8)
			binary.BigEndian.PutUint64(sequenceNumBytes, 1)
			assert.False(t, inputFilter.Test(sequenceNumBytes))
			sequenceNumBytes = make([]byte, 8)
			binary.BigEndian.PutUint64(sequenceNumBytes, 2)
			assert.True(t, inputFilter.Test(sequenceNumBytes))
			close(keepAliveDone)
			return &multiraftv1.KeepAliveResponse{
				Headers: &multiraftv1.PartitionResponseHeaders{
					Index: 4,
				},
			}, nil
		})
	select {
	case <-keepAliveDone:
	case <-time.After(time.Minute):
		t.FailNow()
	}

	go func() {
		sendResponseCh <- struct{}{}
	}()
	response, err = stream.Recv()
	assert.NoError(t, err)
	assert.NotNil(t, response)

	close(sendResponseCh)

	keepAliveDone = make(chan struct{})
	partitionServer.EXPECT().KeepAlive(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, request *multiraftv1.KeepAliveRequest) (*multiraftv1.KeepAliveResponse, error) {
			assert.Equal(t, multiraftv1.PartitionID(1), request.Headers.PartitionID)
			assert.Equal(t, multiraftv1.SessionID(1), request.SessionID)
			assert.Equal(t, multiraftv1.SequenceNum(2), request.LastInputSequenceNum)
			inputFilter := &bloom.BloomFilter{}
			assert.NoError(t, json.Unmarshal(request.InputFilter, inputFilter))
			sequenceNumBytes := make([]byte, 8)
			binary.BigEndian.PutUint64(sequenceNumBytes, 1)
			assert.False(t, inputFilter.Test(sequenceNumBytes))
			sequenceNumBytes = make([]byte, 8)
			binary.BigEndian.PutUint64(sequenceNumBytes, 2)
			assert.True(t, inputFilter.Test(sequenceNumBytes))
			close(keepAliveDone)
			return &multiraftv1.KeepAliveResponse{
				Headers: &multiraftv1.PartitionResponseHeaders{
					Index: 5,
				},
			}, nil
		})
	select {
	case <-keepAliveDone:
	case <-time.After(time.Minute):
		t.FailNow()
	}

	response, err = stream.Recv()
	assert.Equal(t, io.EOF, err)
	assert.Nil(t, response)

	keepAliveDone = make(chan struct{})
	partitionServer.EXPECT().KeepAlive(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, request *multiraftv1.KeepAliveRequest) (*multiraftv1.KeepAliveResponse, error) {
			assert.Equal(t, multiraftv1.PartitionID(1), request.Headers.PartitionID)
			assert.Equal(t, multiraftv1.SessionID(1), request.SessionID)
			assert.Equal(t, multiraftv1.SequenceNum(2), request.LastInputSequenceNum)
			inputFilter := &bloom.BloomFilter{}
			assert.NoError(t, json.Unmarshal(request.InputFilter, inputFilter))
			sequenceNumBytes := make([]byte, 8)
			binary.BigEndian.PutUint64(sequenceNumBytes, 1)
			assert.False(t, inputFilter.Test(sequenceNumBytes))
			sequenceNumBytes = make([]byte, 8)
			binary.BigEndian.PutUint64(sequenceNumBytes, 2)
			assert.False(t, inputFilter.Test(sequenceNumBytes))
			_, ok := request.LastOutputSequenceNums[2]
			assert.False(t, ok)
			close(keepAliveDone)
			return &multiraftv1.KeepAliveResponse{
				Headers: &multiraftv1.PartitionResponseHeaders{
					Index: 6,
				},
			}, nil
		})
	select {
	case <-keepAliveDone:
	case <-time.After(time.Minute):
		t.FailNow()
	}

	partitionServer.EXPECT().CloseSession(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, request *multiraftv1.CloseSessionRequest) (*multiraftv1.CloseSessionResponse, error) {
			return &multiraftv1.CloseSessionResponse{
				Headers: &multiraftv1.PartitionResponseHeaders{
					Index: 7,
				},
			}, nil
		})
	err = client.Close(context.TODO())
	assert.NoError(t, err)
}

func TestStreamQueryCancel(t *testing.T) {
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
	timeout := 10 * time.Second
	err = client.Connect(context.TODO(), multiraftv1.DriverConfig{
		Partitions: []multiraftv1.PartitionConfig{
			{
				PartitionID: 1,
				Leader:      "localhost:5678",
			},
		},
		SessionTimeout: &timeout,
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

	query := StreamQuery[*TestQueryResponse](primitive)
	testServer.EXPECT().TestStreamQuery(gomock.Any(), gomock.Any()).
		Return(errors.ToProto(errors.NewUnavailable("unavailable")))
	testServer.EXPECT().TestStreamQuery(gomock.Any(), gomock.Any()).
		DoAndReturn(func(request *TestQueryRequest, stream Test_TestStreamQueryServer) error {
			assert.Equal(t, multiraftv1.PartitionID(1), request.Headers.PartitionID)
			assert.Equal(t, multiraftv1.SessionID(1), request.Headers.SessionID)
			assert.Equal(t, multiraftv1.SequenceNum(2), request.Headers.SequenceNum)
			assert.Equal(t, multiraftv1.PrimitiveID(2), request.Headers.PrimitiveID)
			assert.Nil(t, stream.Send(&TestQueryResponse{
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
			}))
			assert.Nil(t, stream.Send(&TestQueryResponse{
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
			}))
			<-stream.Context().Done()
			return stream.Context().Err()
		})

	ctx, cancel := context.WithCancel(context.Background())
	stream, err := query.Run(func(conn *grpc.ClientConn, headers *multiraftv1.QueryRequestHeaders) (QueryStream[*TestQueryResponse], error) {
		return NewTestClient(conn).TestStreamQuery(ctx, &TestQueryRequest{
			Headers: headers,
		})
	})
	assert.NoError(t, err)

	response, err := stream.Recv()
	assert.NoError(t, err)
	assert.NotNil(t, response)
	response, err = stream.Recv()
	assert.NoError(t, err)
	assert.NotNil(t, response)

	keepAliveDone := make(chan struct{})
	partitionServer.EXPECT().KeepAlive(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, request *multiraftv1.KeepAliveRequest) (*multiraftv1.KeepAliveResponse, error) {
			assert.Equal(t, multiraftv1.PartitionID(1), request.Headers.PartitionID)
			assert.Equal(t, multiraftv1.SessionID(1), request.SessionID)
			assert.Equal(t, multiraftv1.SequenceNum(2), request.LastInputSequenceNum)
			inputFilter := &bloom.BloomFilter{}
			assert.NoError(t, json.Unmarshal(request.InputFilter, inputFilter))
			sequenceNumBytes := make([]byte, 8)
			binary.BigEndian.PutUint64(sequenceNumBytes, 1)
			assert.False(t, inputFilter.Test(sequenceNumBytes))
			sequenceNumBytes = make([]byte, 8)
			binary.BigEndian.PutUint64(sequenceNumBytes, 2)
			assert.True(t, inputFilter.Test(sequenceNumBytes))
			close(keepAliveDone)
			return &multiraftv1.KeepAliveResponse{
				Headers: &multiraftv1.PartitionResponseHeaders{
					Index: 4,
				},
			}, nil
		})
	select {
	case <-keepAliveDone:
	case <-time.After(time.Minute):
		t.FailNow()
	}

	cancel()

	keepAliveDone = make(chan struct{})
	partitionServer.EXPECT().KeepAlive(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, request *multiraftv1.KeepAliveRequest) (*multiraftv1.KeepAliveResponse, error) {
			assert.Equal(t, multiraftv1.PartitionID(1), request.Headers.PartitionID)
			assert.Equal(t, multiraftv1.SessionID(1), request.SessionID)
			assert.Equal(t, multiraftv1.SequenceNum(2), request.LastInputSequenceNum)
			inputFilter := &bloom.BloomFilter{}
			assert.NoError(t, json.Unmarshal(request.InputFilter, inputFilter))
			sequenceNumBytes := make([]byte, 8)
			binary.BigEndian.PutUint64(sequenceNumBytes, 1)
			assert.False(t, inputFilter.Test(sequenceNumBytes))
			sequenceNumBytes = make([]byte, 8)
			binary.BigEndian.PutUint64(sequenceNumBytes, 2)
			assert.True(t, inputFilter.Test(sequenceNumBytes))
			close(keepAliveDone)
			return &multiraftv1.KeepAliveResponse{
				Headers: &multiraftv1.PartitionResponseHeaders{
					Index: 5,
				},
			}, nil
		})
	select {
	case <-keepAliveDone:
	case <-time.After(time.Minute):
		t.FailNow()
	}

	_, err = stream.Recv()
	assert.Error(t, err)

	keepAliveDone = make(chan struct{})
	partitionServer.EXPECT().KeepAlive(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, request *multiraftv1.KeepAliveRequest) (*multiraftv1.KeepAliveResponse, error) {
			assert.Equal(t, multiraftv1.PartitionID(1), request.Headers.PartitionID)
			assert.Equal(t, multiraftv1.SessionID(1), request.SessionID)
			assert.Equal(t, multiraftv1.SequenceNum(2), request.LastInputSequenceNum)
			inputFilter := &bloom.BloomFilter{}
			assert.NoError(t, json.Unmarshal(request.InputFilter, inputFilter))
			sequenceNumBytes := make([]byte, 8)
			binary.BigEndian.PutUint64(sequenceNumBytes, 1)
			assert.False(t, inputFilter.Test(sequenceNumBytes))
			sequenceNumBytes = make([]byte, 8)
			binary.BigEndian.PutUint64(sequenceNumBytes, 2)
			assert.False(t, inputFilter.Test(sequenceNumBytes))
			close(keepAliveDone)
			return &multiraftv1.KeepAliveResponse{
				Headers: &multiraftv1.PartitionResponseHeaders{
					Index: 5,
				},
			}, nil
		})
	select {
	case <-keepAliveDone:
	case <-time.After(time.Minute):
		t.FailNow()
	}

	partitionServer.EXPECT().CloseSession(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, request *multiraftv1.CloseSessionRequest) (*multiraftv1.CloseSessionResponse, error) {
			return &multiraftv1.CloseSessionResponse{
				Headers: &multiraftv1.PartitionResponseHeaders{
					Index: 6,
				},
			}, nil
		})
	err = client.Close(context.TODO())
	assert.NoError(t, err)
}
