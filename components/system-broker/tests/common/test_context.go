package common

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/gavv/httpexpect"
	"github.com/kyma-incubator/compass/components/system-broker/internal/config"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/env"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/server"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/uuid"
	"k8s.io/apimachinery/pkg/util/wait"
)

const SystemBrokerServer = "system-broker-server"
const DirectorServer = "director-server"

type TestContext struct {
	SystemBroker *httpexpect.Expect

	servers map[string]FakeServer
}

func (tc *TestContext) CleanUp() {
	for _, server := range tc.servers {
		server.Close()
	}
}

type TestContextBuilder struct {
	envHooks []func(env env.Environment, servers map[string]FakeServer)

	Environment env.Environment

	Servers    map[string]FakeServer
	HttpClient *http.Client
}

// NewTestContextBuilder sets up a builder with default values
func NewTestContextBuilder() *TestContextBuilder {
	return &TestContextBuilder{
		Environment: TestEnv(),
		envHooks: []func(env env.Environment, servers map[string]FakeServer){
			func(env env.Environment, servers map[string]FakeServer) {
				env.Set("server.shutdown_timeout", "1s")
				port := findFreePort()
				env.Set("server.port", port)
				env.Set("server.self_url", "http://localhost:"+port)
			},
		},
		Servers: map[string]FakeServer{},
		HttpClient: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   20 * time.Second,
					KeepAlive: 20 * time.Second,
				}).DialContext,
				MaxIdleConns:          100,
				IdleConnTimeout:       30 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},
	}
}

func (tcb *TestContextBuilder) WithDefaultEnv(env env.Environment) *TestContextBuilder {
	tcb.Environment = env

	return tcb
}

func (tcb *TestContextBuilder) WithEnvExtensions(fs ...func(e env.Environment, servers map[string]FakeServer)) *TestContextBuilder {
	tcb.envHooks = append(tcb.envHooks, fs...)

	return tcb
}

func (tcb *TestContextBuilder) WithHttpClient(client *http.Client) *TestContextBuilder {
	tcb.HttpClient = client

	return tcb
}

func (tcb *TestContextBuilder) Build(t *testing.T) *TestContext {
	for _, envPostHook := range tcb.envHooks {
		envPostHook(tcb.Environment, tcb.Servers)
	}

	sbServer := newSystemBrokerServer(tcb.Environment)
	tcb.Servers[SystemBrokerServer] = sbServer

	gqlMockHandler, err := NewGqlFakeRouter("director", "../../../director/pkg/graphql/schema.graphql")
	if err != nil {
		panic(fmt.Errorf("could not build gql mock handler: %s", err))
	}
	gqlMockServer := NewGqlFakeServer(gqlMockHandler.Handler())
	tcb.Servers[DirectorServer] = gqlMockServer

	systemBroker := httpexpect.New(t, sbServer.URL()).Builder(func(request *httpexpect.Request) {
		request.WithClient(tcb.HttpClient)
	})

	testContext := &TestContext{
		SystemBroker: systemBroker,
		servers:      tcb.Servers,
	}

	return testContext
}

func TestEnv() env.Environment {
	env, err := env.Default(context.TODO())
	if err != nil {
		panic(err)
	}
	return env
}

type testSystemBrokerServer struct {
	url             string
	cancel          context.CancelFunc
	shutdownTimeout time.Duration
}

func (ts *testSystemBrokerServer) URL() string {
	return ts.url
}

func (ts *testSystemBrokerServer) Close() {
	ts.cancel()
	time.Sleep(ts.shutdownTimeout)
}

func newSystemBrokerServer(sbEnv env.Environment) FakeServer {
	ctx, cancel := context.WithCancel(context.Background())

	cfg, err := config.New(sbEnv)
	if err != nil {
		panic(err)
	}

	sbServer := server.New(cfg.Server, uuid.NewService())

	sbServer.Addr = "localhost:" + strconv.Itoa(cfg.Server.Port) // Needed to avoid annoying macOS permissions popup

	go sbServer.Start(ctx)

	err = wait.PollImmediate(time.Millisecond*250, time.Second*5, func() (bool, error) {
		_, err := http.Get(fmt.Sprintf("http://%s", sbServer.Addr))
		if err != nil {
			log.Printf("Waiting for server to start: %v", err)
			return false, nil
		}
		return true, nil
	})

	if err != nil {
		panic(err)
	}

	return &testSystemBrokerServer{
		url:             cfg.Server.SelfURL + cfg.Server.RootAPI,
		cancel:          cancel,
		shutdownTimeout: cfg.Server.ShutdownTimeout,
	}
}

func findFreePort() string {
	// Create a new listener without specifying a port which will result in an open port being chosen
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}

	defer listener.Close()

	hostString := listener.Addr().String()
	_, port, err := net.SplitHostPort(hostString)
	if err != nil {
		panic(err)
	}

	return port
}
