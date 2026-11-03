//go:build !windows

package pty

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"sync"

	creackpty "github.com/creack/pty"
)

type linuxSession struct {
	cmd  *exec.Cmd
	ptmx *os.File
	once sync.Once
}

func StartSession(username string, cols, rows uint16) (Session, error) {
	var cmd *exec.Cmd

	targetUser := username
	if targetUser == "" {
		targetUser = "root"
	}

	// Resolve target user's home directory
	var homeDir string
	if u, err := user.Lookup(targetUser); err == nil && u.HomeDir != "" {
		homeDir = u.HomeDir
	} else if targetUser == "root" {
		homeDir = "/root"
	} else {
		homeDir = "/home/" + targetUser
	}

	if targetUser != "root" {
		// runuser -l simulates a full login, cd-ing into the target user's home directory
		cmd = exec.Command("runuser", "-l", targetUser, "-c", "exec /bin/bash -l")
	} else {
		cmd = exec.Command("/bin/bash", "-l")
	}

	// Explicitly set the initial process working directory to user's home directory
	if _, statErr := os.Stat(homeDir); statErr == nil {
		cmd.Dir = homeDir
	}

	cmd.Env = append(os.Environ(),
		"TERM=xterm-256color",
		"COLORTERM=truecolor",
		"LANG=en_US.UTF-8",
		"HOME="+homeDir,
		"USER="+targetUser,
	)

	ws := &creackpty.Winsize{
		Cols: cols,
		Rows: rows,
	}

	ptmx, err := creackpty.StartWithSize(cmd, ws)
	if err != nil {
		return nil, fmt.Errorf("failed to start pty session: %w", err)
	}

	return &linuxSession{
		cmd:  cmd,
		ptmx: ptmx,
	}, nil
}

func (s *linuxSession) Read(p []byte) (n int, err error) {
	return s.ptmx.Read(p)
}

func (s *linuxSession) Write(p []byte) (n int, err error) {
	return s.ptmx.Write(p)
}

func (s *linuxSession) Resize(cols, rows uint16) error {
	ws := &creackpty.Winsize{
		Cols: cols,
		Rows: rows,
	}
	return creackpty.Setsize(s.ptmx, ws)
}

func (s *linuxSession) Close() error {
	var closeErr error
	s.once.Do(func() {
		if s.cmd != nil && s.cmd.Process != nil {
			if err := s.cmd.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
				// Process might already be terminated
			}
			if err := s.cmd.Wait(); err != nil {
				// Process termination wait completed
			}
		}
		if s.ptmx != nil {
			closeErr = s.ptmx.Close()
		}
	})
	return closeErr
}
