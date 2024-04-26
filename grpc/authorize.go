package grpc

import (
	"context"
	"fmt"

	pb "vacancies/grpc/authorize_proto"

	"google.golang.org/grpc"
)

// AuthorizationData - данные для авторизации на сайте
type AuthorizationData struct {
	Cookies string
}

// GetAuthorizationData - получение данных авторизации
func GetAuthorizationData(target string) (AuthorizationData, error) {
	conn, err := grpc.Dial(target, grpc.WithInsecure())
	if err != nil {
		return AuthorizationData{}, fmt.Errorf("grpc.Dial: %w", err)
	}
	defer conn.Close()

	grpcClient := pb.NewAuthorizationServiceClient(conn)
	response, err := grpcClient.GetAuthorizationData(context.Background(), &pb.AuthorizationDataRequest{})
	if err != nil {
		return AuthorizationData{}, fmt.Errorf("client.GetAuthorizationData: %v", err)
	}

	return AuthorizationData{Cookies: response.Cookies}, nil
}
