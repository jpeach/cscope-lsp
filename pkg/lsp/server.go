package lsp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"

	"github.com/jpeach/cscope-cquery/pkg/lsp/cquery"

	"github.com/sourcegraph/jsonrpc2"
)

// ServerOpts ...
type ServerOpts struct {
	// Path to the LSP server executable.
	Path string

	// Args are additional arguments passed to the LSP server at launch.
	Args []string
}

// NewServer ...
func NewServer() (*Server, error) {
	return &Server{
		lock: &sync.Mutex{},
	}, nil
}

type handler struct {
}

func (h *handler) Handle(ctx context.Context, c *jsonrpc2.Conn, r *jsonrpc2.Request) {
	switch r.Method {
	case "$cquery/progress":
		var p cquery.Progress
		if err := json.Unmarshal(*r.Params, &p); err != nil {
			log.Printf("failed to unmarshall %s message: %s",
				r.Method, err)
		}
	default:
		log.Printf("handler called for %s\n", r.Method)
	}
}

type rwc struct {
	stdin  io.WriteCloser
	stdout io.ReadCloser
}

func (r *rwc) Read(p []byte) (int, error) {
	return r.stdout.Read(p)
}

func (r *rwc) Write(p []byte) (int, error) {
	return r.stdin.Write(p)
}

func (r *rwc) Close() error {
	// Don't close, because really the Server owns these files.
	return nil
}

// Server is an instance of a LSP server process.
type Server struct {
	cmd  *exec.Cmd
	lock *sync.Mutex

	conn *jsonrpc2.Conn

	in  io.WriteCloser
	out io.ReadCloser
}

func (s *Server) rwc() *rwc {
	return &rwc{
		stdin:  s.in,
		stdout: s.out,
	}
}

func (s *Server) reset() {
	if s.in != nil {
		s.in.Close()
		s.in = nil
	}

	if s.out != nil {
		s.out.Close()
		s.out = nil
	}

	s.cmd = nil
}

func (s *Server) start(path string, args []string) error {
	var err error

	s.cmd = exec.Command(path, args...)
	s.cmd.Stderr = os.Stderr

	// TODO(jpeach): Set up the child death sig on Linux.
	/*
		s.cmd.SysProcAttr = &syscall.SysProcAttr{
			Pdeathsig: os.SIGKILL,
		}
	*/

	s.in, err = s.cmd.StdinPipe()
	if err != nil {
		s.reset()
		return err
	}

	s.out, err = s.cmd.StdoutPipe()
	if err != nil {
		s.reset()
		return err
	}

	err = s.cmd.Start()
	if err != nil {
		s.reset()
		return err
	}

	rpcOpt := []jsonrpc2.ConnOpt{}

	s.conn = jsonrpc2.NewConn(
		context.Background(),
		jsonrpc2.NewBufferedStream(s.rwc(), jsonrpc2.VSCodeObjectCodec{}),
		&handler{},
		rpcOpt...)

	return nil
}

// Start ...
func (s *Server) Start(opts *ServerOpts) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.cmd != nil {
		return errors.New("server already running")
	}

	if err := s.start(opts.Path, opts.Args); err != nil {
		return err
	}

	go func() {
		s.cmd.Wait()

		s.lock.Lock()
		defer s.lock.Unlock()

		s.reset()

		// TODO(jpeach): Send a notification that the server died and
		// the caller should reinitialize it.
	}()

	return nil
}

// Call ...
func (s *Server) Call(ctx context.Context, method string, params interface{}, result interface{}) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.conn.Call(ctx, method, params, result)
}

// Notify ...
func (s *Server) Notify(ctx context.Context, method string, params interface{}) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.conn.Notify(context.Background(), method, &params)
}
