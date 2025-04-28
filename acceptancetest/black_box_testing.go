package acceptancetest

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

const baseBinName = "test_binary"

func RunServer(path string, port string) (cleanup, interrupt func() error, err error) {
	binPath, err := buildBin(path)
	if err != nil {
		return nil, nil, fmt.Errorf("error when building binary: %w", err)
	}

	interrupt, kill, err := runBin(binPath)
	if err != nil {
		return nil, nil, fmt.Errorf("error when running binary %w", err)
	}

	if err := waitForServerToListen(port); err != nil {
		kill()
		return nil, nil, err
	}

	cleanup = func() error {
		if err := kill(); err != nil {
			return nil
		}
		return os.Remove(binPath)
	}

	return cleanup, interrupt, nil
}

func buildBin(path string) (binPath string, err error) {
	binName := baseBinName
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}

	buildCmd := exec.Command("go", "build", "-o", binName, path)
	if err := buildCmd.Run(); err != nil {
		return "", err
	}

	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, binName), nil
}

func runBin(binPath string) (interrupt, kill func() error, err error) {
	cmd := exec.Command(binPath)
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, nil, err
	}

	interrupt = func() error {
		return cmd.Process.Signal(os.Interrupt)
	}
	kill = func() error {
		return cmd.Process.Kill()
	}

	return interrupt, kill, nil
}

func waitForServerToListen(port string) error {
	for range 20 {
		conn, _ := net.Dial("tcp", net.JoinHostPort("localhost", port))
		if conn != nil {
			conn.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return errors.New("timeout waiting for server to listen")
}
