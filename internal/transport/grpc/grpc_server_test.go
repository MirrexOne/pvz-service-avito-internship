package grpc_test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	pb "pvz-service-avito-internship/pkg/grpc/pvz/v1"
	"testing"
)

type MockPVZService struct {
	mock.Mock
	pb.UnimplementedPVZServiceServer
}

func (m *MockPVZService) GetPVZList(ctx context.Context, req *pb.GetPVZListRequest) (*pb.GetPVZListResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*pb.GetPVZListResponse), args.Error(1)
}

func setupTestServer(mockService *MockPVZService) (*grpc.Server, string) {
	server := grpc.NewServer()
	pb.RegisterPVZServiceServer(server, mockService)

	lis, _ := net.Listen("tcp", ":0")
	go server.Serve(lis)

	return server, lis.Addr().String()
}

func TestGetPVZList(t *testing.T) {
	mockService := new(MockPVZService)
	server, addr := setupTestServer(mockService)
	defer server.Stop()

	conn, _ := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	defer conn.Close()

	client := pb.NewPVZServiceClient(conn)

	t.Run("Successful request", func(t *testing.T) {
		mockService.On("GetPVZList", mock.Anything, &pb.GetPVZListRequest{}).
			Return(&pb.GetPVZListResponse{
				Pvzs: []*pb.PVZ{
					{Id: "pvz-001", City: "Москва"},
				},
			}, nil)

		resp, err := client.GetPVZList(context.Background(), &pb.GetPVZListRequest{})
		assert.NoError(t, err)
		assert.Len(t, resp.Pvzs, 1)
		assert.Equal(t, "Москва", resp.Pvzs[0].City)

		mockService.AssertExpectations(t)
	})
}
