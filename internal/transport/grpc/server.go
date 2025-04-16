package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"net" // Пакет для сетевых операций (Listen)

	// Используем стандартные пакеты gRPC
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"                       // Стандартные коды ошибок gRPC
	"google.golang.org/grpc/status"                      // Для создания gRPC ошибок со статусом
	"google.golang.org/protobuf/types/known/timestamppb" // Для конвертации time.Time

	// Используем относительные пути в рамках проекта
	"pvz-service-avito-internship/internal/domain"
	// Импортируем сгенерированный Protobuf/gRPC код
	pb "pvz-service-avito-internship/pkg/grpc/pvz/v1"
	// Импортируем middleware для получения RequestID (хотя здесь он не сильно нужен, но для консистентности)
	"pvz-service-avito-internship/internal/middleware"
)

// Server структура gRPC сервера.
type Server struct {
	// Встраиваем UnimplementedPVZServiceServer для обратной совместимости.
	// Если в будущем в .proto добавятся новые методы, сервер не сломается,
	// а будет возвращать ошибку Unimplemented для этих методов.
	pb.UnimplementedPVZServiceServer

	log      *slog.Logger         // Логгер
	pvzRepo  domain.PVZRepository // Зависимость от репозитория ПВЗ для получения данных
	grpcServ *grpc.Server         // Экземпляр gRPC сервера из пакета google.golang.org/grpc
	port     string               // Порт, на котором будет слушать сервер
	lis      net.Listener         // Сетевой слушатель (для graceful shutdown)
}

// NewServer создает и конфигурирует новый экземпляр gRPC сервера.
func NewServer(log *slog.Logger, pvzRepo domain.PVZRepository, port string) *Server {
	// Создаем новый gRPC сервер. Здесь можно добавить опции, например, interceptors.
	grpcServerInstance := grpc.NewServer(
	// Пример добавления Unary Interceptor для логирования или метрик:
	// grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
	//    grpc_recovery.UnaryServerInterceptor(), // Перехват паник
	//    grpc_slog.UnaryServerInterceptor(log),   // Логирование запросов (из go-grpc-middleware/providers/slog)
	//    grpc_prometheus.UnaryServerInterceptor, // Метрики (из go-grpc-middleware/providers/prometheus)
	// )),
	)

	s := &Server{
		log:      log,
		pvzRepo:  pvzRepo,
		grpcServ: grpcServerInstance,
		port:     port,
	}

	// Регистрируем нашу реализацию Server как обработчик для PVZService.
	pb.RegisterPVZServiceServer(grpcServerInstance, s)

	return s
}

// GetPVZList реализует gRPC метод PVZService.GetPVZList.
// Получает все ПВЗ из репозитория и возвращает их в формате Protobuf сообщений.
func (s *Server) GetPVZList(ctx context.Context, req *pb.GetPVZListRequest) (*pb.GetPVZListResponse, error) {
	const op = "GRPCServer.GetPVZList"
	// Получаем Request ID из контекста, если он был передан (например, через gRPC metadata)
	// В данном случае он не передается, но оставим для примера
	reqID := middleware.GetRequestIDFromContext(ctx)
	log := s.log.With(slog.String("op", op), slog.String("request_id", reqID))

	log.Info("Received GetPVZList request")

	// Вызываем метод репозитория для получения *всех* ПВЗ
	pvzs, err := s.pvzRepo.ListAll(ctx)
	if err != nil {
		log.Error("Failed to list all PVZs from repository", slog.String("error", err.Error()))
		// Возвращаем стандартизированную gRPC ошибку Internal
		// Оборачиваем оригинальную ошибку репозитория, но не отдаем ее напрямую клиенту.
		return nil, status.Errorf(codes.Internal, "failed to retrieve PVZ list")
	}

	log.Info("PVZ list retrieved successfully from repository", slog.Int("count", len(pvzs)))

	// Конвертируем доменные структуры []domain.PVZ в Protobuf сообщения []*pb.PVZ
	respPVZs := make([]*pb.PVZ, 0, len(pvzs))
	for _, p := range pvzs {
		respPVZs = append(respPVZs, &pb.PVZ{
			Id:               p.ID.String(),                             // Конвертируем uuid.UUID в string
			RegistrationDate: timestamppb.New(p.RegistrationDate.UTC()), // Конвертируем time.Time в google.protobuf.Timestamp (важно использовать UTC)
			City:             string(p.City),                            // Конвертируем domain.City в string
		})
	}

	// Создаем и возвращаем ответное сообщение
	response := &pb.GetPVZListResponse{
		Pvzs: respPVZs, // Используем имя поля из .proto файла (было pvzs, исправлено на Pvz)
	}

	log.Info("Successfully prepared GetPVZList response", slog.Int("count", len(response.Pvzs)))
	return response, nil
}

// Start запускает gRPC сервер на прослушивание указанного порта.
// Эта функция блокирующая и вернет ошибку, если не удалось начать прослушивание или сервер упал.
func (s *Server) Start() error {
	const op = "GRPCServer.Start"
	log := s.log.With(slog.String("op", op))

	// Создаем сетевой слушатель (listener) на указанном порту
	address := ":" + s.port
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Error("Failed to listen on gRPC port", slog.String("address", address), slog.String("error", err.Error()))
		return fmt.Errorf("failed to listen on gRPC port %s: %w", s.port, err)
	}
	s.lis = lis // Сохраняем слушатель для GracefulStop

	log.Info("Starting gRPC server listener", slog.String("address", lis.Addr().String()))

	// Запускаем gRPC сервер с созданным слушателем.
	// Эта функция будет блокировать выполнение до тех пор, пока сервер не будет остановлен.
	if err := s.grpcServ.Serve(lis); err != nil {
		// Ошибка Serve обычно возникает при штатной остановке или реальной проблеме
		log.Error("gRPC server Serve failed", slog.String("error", err.Error()))
		return fmt.Errorf("gRPC server failed to serve: %w", err)
	}

	// Этот код достигнется только после остановки сервера
	log.Info("gRPC server has stopped serving")
	return nil
}

// Stop грациозно останавливает gRPC сервер.
// Позволяет текущим запросам завершиться перед полной остановкой.
func (s *Server) Stop() {
	const op = "GRPCServer.Stop"
	log := s.log.With(slog.String("op", op))
	log.Info("Stopping gRPC server gracefully...")

	// Вызываем GracefulStop(), который сначала прекратит принимать новые соединения,
	// затем дождется завершения текущих RPC (или таймаута), и только потом остановит сервер.
	s.grpcServ.GracefulStop()

	// Закрываем слушатель, если он был открыт (на всякий случай)
	if s.lis != nil {
		_ = s.lis.Close() // Игнорируем ошибку закрытия
	}

	log.Info("gRPC server stopped")
}
