package grpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/jinzhu/gorm"
	"github.com/rcttavares/codepix/application/grpc/pb"
	"github.com/rcttavares/codepix/application/usecase"
	"github.com/rcttavares/codepix/infrastructure/repository"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

// authInterceptor requires every unary call to carry a valid "authorization"
// metadata entry matching the grpcAuthToken configured for this service.
func authInterceptor(expectedToken string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		tokens := md.Get("authorization")
		if len(tokens) == 0 || tokens[0] != expectedToken {
			return nil, status.Error(codes.Unauthenticated, "invalid or missing token")
		}

		return handler(ctx, req)
	}
}

func StartGrpcServer(database *gorm.DB, port int) {
	expectedToken := os.Getenv("grpcAuthToken")
	if expectedToken == "" {
		log.Fatal("grpcAuthToken must be set to start the gRPC server")
	}

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(authInterceptor(expectedToken)))

	if os.Getenv("debug") == "true" {
		reflection.Register(grpcServer)
	}

	pixRepository := repository.PixKeyRepositoryDb{Db: database}
	pixUseCase := usecase.PixUseCase{PixKeyRepository: pixRepository}
	pixGrpcService := NewPixGrpcService(pixUseCase)
	pb.RegisterPixServiceServer(grpcServer, pixGrpcService)

	address := fmt.Sprintf("0.0.0.0:%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("cannot start grpc server", err)
	}

	log.Printf("gRPC server has been started on port %d", port)
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("cannot start grpc server", err)
	}
}
