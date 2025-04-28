package acceptancetest

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

const binPrefix = "test_binary"

func BuildBinary(path, name string) (cleanup func() error, binPath string, err error) {
	binName := binPrefix + "_" + name
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}

	buildCmd := exec.Command("go", "build", "-o", binName, path)
	if err := buildCmd.Run(); err != nil {
		return nil, "", fmt.Errorf("cannot build binary due to error: %w", err)
	}

	dir, err := os.Getwd()
	if err != nil {
		return nil, "", err
	}

	binPath = filepath.Join(dir, binName)

	cleanup = func() error {
		return os.Remove(binPath)
	}

	return cleanup, binPath, nil
}

func RunBin(binPath string) (interrupt func() error, err error) {
	cmd := exec.Command(binPath)
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("error running bin path due to: %w", err)
	}

	interrupt = func() error {
		return cmd.Process.Signal(os.Interrupt)
	}

	return interrupt, nil
}

func WaitForServerToListen(port string) {
	for range 20 {
		conn, _ := net.Dial("tcp", net.JoinHostPort("localhost", port))
		if conn != nil {
			conn.Close()
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
}
