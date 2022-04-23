package vitess

import (
	"context"
	"io"
	stdlog "log"
	"os"
	"time"

	"google.golang.org/grpc"

	"github.com/pkg/errors"
	"vitess.io/vitess/go/vt/grpcclient"
	"vitess.io/vitess/go/vt/logutil"
	logutilpb "vitess.io/vitess/go/vt/proto/logutil"
	vtctldatapb "vitess.io/vitess/go/vt/proto/vtctldata"
	vtctlservicepb "vitess.io/vitess/go/vt/proto/vtctlservice"
	"vitess.io/vitess/go/vt/vtctl/grpcclientcommon"
	"vitess.io/vitess/go/vt/vtctl/vtctlclient"
)

type SchemaLoader interface {
	Migrate(dsn DSN, migrationsDir, dbName string) error
}

type schemaLoader struct {
	logger *stdlog.Logger
}

func NewSchemaLoader(logger *stdlog.Logger) SchemaLoader {
	return &schemaLoader{logger}
}

func (s *schemaLoader) Migrate(dsn DSN, vSchemaPath, dbName string) (err error) {
	vSchemaFile, err := os.Open(vSchemaPath)
	if err != nil {
		return errors.WithStack(err)
	}
	defer vSchemaFile.Close()

	byteValue, _ := io.ReadAll(vSchemaFile)
	args := []string{"ApplyVSchema", "--vschema", string(byteValue), dbName}

	vtctlclient.RegisterFactory("grpc", gRPCVtctlClientFactory)
	err = vtctlclient.RunCommandAndWait(context.Background(), dsn.String(), args, func(e *logutilpb.Event) {
		s.logger.Println(e)
	})
	return errors.WithStack(err)
}

type gRPCVtctlClient struct {
	cc *grpc.ClientConn
	c  vtctlservicepb.VtctlClient
}

func gRPCVtctlClientFactory(addr string) (vtctlclient.VtctlClient, error) {
	opt, err := grpcclientcommon.SecureDialOption()
	if err != nil {
		return nil, err
	}
	// create the RPC client
	cc, err := grpcclient.Dial(addr, grpcclient.FailFast(false), opt)
	if err != nil {
		return nil, err
	}
	c := vtctlservicepb.NewVtctlClient(cc)

	return &gRPCVtctlClient{
		cc: cc,
		c:  c,
	}, nil
}

type eventStreamAdapter struct {
	stream vtctlservicepb.Vtctl_ExecuteVtctlCommandClient
}

func (e *eventStreamAdapter) Recv() (*logutilpb.Event, error) {
	le, err := e.stream.Recv()
	if err != nil {
		return nil, err
	}
	return le.Event, nil
}

// ExecuteVtctlCommand is part of the VtctlClient interface
func (client *gRPCVtctlClient) ExecuteVtctlCommand(ctx context.Context, args []string, actionTimeout time.Duration) (logutil.EventStream, error) {
	query := &vtctldatapb.ExecuteVtctlCommandRequest{
		Args:          args,
		ActionTimeout: actionTimeout.Nanoseconds(),
	}

	stream, err := client.c.ExecuteVtctlCommand(ctx, query)
	if err != nil {
		return nil, err
	}
	return &eventStreamAdapter{stream}, nil
}

// Close is part of the VtctlClient interface
func (client *gRPCVtctlClient) Close() {
	client.cc.Close()
}
