package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"pvz-service-avito-internship/internal/domain"
	"pvz-service-avito-internship/internal/middleware"
	pb "pvz-service-avito-internship/pkg/grpc/pvz/v1"
)

type Server struct {
	pb.UnimplementedPVZServiceServer

	log      *slog.Logger
	pvzRepo  domain.PVZRepository
	grpcServ *grpc.Server
	port     string
	lis      net.Listener
}

func NewServer(log *slog.Logger, pvzRepo domain.PVZRepository, port string) *Server {
	grpcServerInstance := grpc.NewServer()

	s := &Server{
		log:      log,
		pvzRepo:  pvzRepo,
		grpcServ: grpcServerInstance,
		port:     port,
	}

	pb.RegisterPVZServiceServer(grpcServerInstance, s)

	return s
}

func (s *Server) GetPVZList(ctx context.Context, req *pb.GetPVZListRequest) (*pb.GetPVZListResponse, error) {
	const op = "GRPCServer.GetPVZList"

	reqID := middleware.GetRequestIDFromContext(ctx)
	log := s.log.With(slog.String("op", op), slog.String("request_id", reqID))

	log.Info("Received GetPVZList request")

	pvzs, err := s.pvzRepo.ListAll(ctx)
	if err != nil {
		log.Error("Failed to list all PVZs from repository", slog.String("error", err.Error()))
		return nil, status.Errorf(codes.Internal, "failed to retrieve PVZ list")
	}

	log.Info("PVZ list retrieved successfully from repository", slog.Int("count", len(pvzs)))

	respPVZs := make([]*pb.PVZ, 0, len(pvzs))
	for _, p := range pvzs {
		respPVZs = append(respPVZs, &pb.PVZ{
			Id:               p.ID.String(),
			RegistrationDate: timestamppb.New(p.RegistrationDate.UTC()),
			City:             string(p.City),
		})
	}

	response := &pb.GetPVZListResponse{
		Pvzs: respPVZs,
	}

	log.Info("Successfully prepared GetPVZList response", slog.Int("count", len(response.Pvzs)))
	return response, nil
}

func (s *Server) Start() error {
	const op = "GRPCServer.Start"
	log := s.log.With(slog.String("op", op))

	address := ":" + s.port
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Error("Failed to listen on gRPC port", slog.String("address", address), slog.String("error", err.Error()))
		return fmt.Errorf("failed to listen on gRPC port %s: %w", s.port, err)
	}
	s.lis = lis

	log.Info("Starting gRPC server listener", slog.String("address", lis.Addr().String()))

	if err := s.grpcServ.Serve(lis); err != nil {
		log.Error("gRPC server Serve failed", slog.String("error", err.Error()))
		return fmt.Errorf("gRPC server failed to serve: %w", err)
	}

	log.Info("gRPC server has stopped serving")
	return nil
}

func (s *Server) Stop() {
	const op = "GRPCServer.Stop"
	log := s.log.With(slog.String("op", op))
	log.Info("Stopping gRPC server gracefully...")

	s.grpcServ.GracefulStop()

	if s.lis != nil {
		_ = s.lis.Close()
	}

	log.Info("gRPC server stopped")
}
