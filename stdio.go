package mcp

import (
	"bufio"
	"context"
	"io"
	"os"
)

// ServeStdio executes the MCP server over standard input and output (stdin/stdout).
// This is used for desktop agents (Claude Desktop, Cursor, local terminal CLI).
func (s *Server) ServeStdio(ctx context.Context, in io.Reader, out io.Writer) error {
	if in == nil {
		in = os.Stdin
	}
	if out == nil {
		out = os.Stdout
	}

	scanner := bufio.NewScanner(in)
	// Allow up to 10MB per line for large data payloads
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 10*1024*1024)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		respBytes, err := s.HandleMessage(ctx, line)
		if err != nil {
			return err
		}

		if len(respBytes) > 0 {
			if _, err := out.Write(append(respBytes, '\n')); err != nil {
				return err
			}
		}
	}

	return scanner.Err()
}
