package main

import (
	"fmt"
	"log"
	"net"

	pb "github.com/kyma-incubator/compass/components/director/protobuf"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type handler struct{}

func (handler) Applications(ctx context.Context, in *pb.ApplicationsInput) (*pb.ApplicationsResponse, error) {
	return &pb.ApplicationsResponse{
		Applications: []*pb.Application{
			{
				Id:             "foo-test",
				Name:           "Test",
				Description:    "Foo",
				HealthCheckURL: "foo.bar",
				Tenant:         "1",
				Annotations: map[string][]byte{
					"test": []byte("string"),
				},
				Status: &pb.ApplicationStatus{
					Condition: pb.ApplicationStatusCondition_INITIAL,
					Timestamp: int64(3231),
				},
			},
			{
				Id:             "bar-test",
				Name:           "Test2",
				Description:    "aa.bb",
				HealthCheckURL: "bar.com",
				Tenant:         "1",
				Annotations: map[string][]byte{
					"test": []byte("string2"),
				},
				Status: &pb.ApplicationStatus{
					Condition: pb.ApplicationStatusCondition_INITIAL,
					Timestamp: int64(3231),
				},
			},
		},
	}, nil
}

func (handler) Apis(ctx context.Context, in *pb.ApplicationRoot) (*pb.ApplicationApisResult, error) {
	return &pb.ApplicationApisResult{
		ApplicationApis: []*pb.ApplicationApi{
			{
				ID:        "1",
				TargetURL: fmt.Sprintf("%s.foo.bar", in.ID),
			},
			{
				ID:        "2",
				TargetURL: fmt.Sprintf("%s.foo.bar", in.ID),
			},
		},
	}, nil
}

func newHandler() *handler {
	return &handler{}
}

func main() {
	log.Println("Starting Director...")
	addr := fmt.Sprintf("127.0.0.1:4000")
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	srv := grpc.NewServer()
	pb.RegisterDirectorServer(srv, newHandler())
	reflection.Register(srv)

	err = srv.Serve(listener)
	if err != nil {
		panic(err)
	}
}
