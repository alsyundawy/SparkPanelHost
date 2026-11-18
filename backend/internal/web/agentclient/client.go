package agentclient

import (
	"fmt"

	agentv1 "github.com/alsyundawy/SparkPanelHost/backend/gen/agent/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	client agentv1.SystemAgentServiceClient
}

func NewClient(socketPath string) (*Client, error) {
	target := "unix://" + socketPath
	conn, err := grpc.NewClient(target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create client for panel-agent on %s: %w", socketPath, err)
	}

	return &Client{
		conn:   conn,
		client: agentv1.NewSystemAgentServiceClient(conn),
	}, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) Agent() agentv1.SystemAgentServiceClient {
	return c.client
}
