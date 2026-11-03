//go:build windows

package pty

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
)

type windowsSession struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	once   sync.Once
}

func StartSession(username string, cols, rows uint16) (Session, error) {
	cmd := exec.Command("cmd.exe")

	if homeDir, err := os.UserHomeDir(); err == nil && homeDir != "" {
		cmd.Dir = homeDir
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to open stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		if closeErr := stdin.Close(); closeErr != nil {
			// ignore cleanup error
		}
		return nil, fmt.Errorf("failed to open stdout pipe: %w", err)
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		if closeErr := stdin.Close(); closeErr != nil {
			// ignore cleanup error
		}
		if closeErr := stdout.Close(); closeErr != nil {
			// ignore cleanup error
		}
		return nil, fmt.Errorf("failed to start windows shell: %w", err)
	}

	return &windowsSession{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
	}, nil
}

func (s *windowsSession) Read(p []byte) (n int, err error) {
	return s.stdout.Read(p)
}

func (s *windowsSession) Write(p []byte) (n int, err error) {
	return s.stdin.Write(p)
}

func (s *windowsSession) Resize(cols, rows uint16) error {
	return nil
}

func (s *windowsSession) Close() error {
	var firstErr error
	s.once.Do(func() {
		if s.stdin != nil {
			if err := s.stdin.Close(); err != nil && firstErr == nil {
				firstErr = err
			}
		}
		if s.stdout != nil {
			if err := s.stdout.Close(); err != nil && firstErr == nil {
				firstErr = err
			}
		}
		if s.cmd != nil && s.cmd.Process != nil {
			if err := s.cmd.Process.Kill(); err != nil && firstErr == nil {
				firstErr = err
			}
			if err := s.cmd.Wait(); err != nil && firstErr == nil {
				// wait completed
			}
		}
	})
	return firstErr
}
