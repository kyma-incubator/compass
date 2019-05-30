package director

import (
	"context"
	"time"

	pb "github.com/kyma-incubator/compass/components/gateway/protobuf"
	grpcpool "github.com/processout/grpc-go-pool"
	"google.golang.org/grpc"
)

const (
	initialOpenSocketNo = 1
	maxOpenSocketNo     = 100
	socketIdleTimeout   = 180 * time.Second
)

type Client struct {
	requestTimeout time.Duration
	connPool       *grpcpool.Pool
}

func NewClient() (*Client, error) {
	pool, err := grpcpool.New(connFactory("127.0.0.1:4000"), initialOpenSocketNo, maxOpenSocketNo, socketIdleTimeout)
	if err != nil {
		return nil, err
	}

	return &Client{
		requestTimeout: 60 * time.Second,
		connPool:       pool,
	}, nil
}

func (d *Client) DirectorClient() (pb.DirectorClient, *grpcpool.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), d.requestTimeout)
	defer cancel()

	var err error
	conn, err := d.connPool.Get(ctx)
	if err != nil {
		return nil, nil, err
	}

	client := pb.NewDirectorClient(conn.ClientConn)

	return client, conn, nil
}

func connFactory(addr string) func() (*grpc.ClientConn, error) {
	return func() (*grpc.ClientConn, error) {
		conn, err := grpc.Dial(addr, grpc.WithInsecure())
		if err != nil {
			return nil, err
		}
		return conn, err
	}
}
