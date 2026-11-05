package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	pb "github.com/anuragbapat10/trust-box.git/internal/gen/trustbox/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type InstanceServer struct {
	logger *slog.Logger
}

func NewInstanceServer(logger *slog.Logger) *InstanceServer {
	return &InstanceServer{
		logger: logger,
	}
}

func (s *InstanceServer) startup(ctx context.Context, id string) (*pb.NodeInstance, error) {
	if id == "" {
		return nil, fmt.Errorf("instance id cannot be empty")
	}

	socketPath := fmt.Sprintf("/tmp/firecracker-%s.socket", id)
	_ = os.Remove(socketPath)

	// 1. Spawn Firecracker Process
	cmd := exec.CommandContext(ctx, "/usr/local/bin/firecracker", "--api-sock", socketPath)
	cmd.Stdin = os.Stdin

	logPath := fmt.Sprintf("/tmp/trust-box/firecracker-%s.log", id)
	_ = os.MkdirAll(filepath.Dir(logPath), 0o755)
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to open firecracker log: %w", err)
	}
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	if err := cmd.Start(); err != nil {
		_ = logFile.Close()
		return nil, fmt.Errorf("failed to start firecracker: %w", err)
	}

	// Wait for the Unix socket to appear
	socketReady := false
	for range 30 {
		if _, err := os.Stat(socketPath); err == nil {
			socketReady = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !socketReady {
		_ = cmd.Process.Kill()
		return nil, fmt.Errorf("firecracker socket %s did not appear in time", socketPath)
	}

	// Configure HTTP client to talk over Unix socket
	httpClient := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
	}

	base := "http://localhost"

	// 2. Configure Machine
	s.logger.InfoContext(ctx, "configuring machine", slog.String("id", id))
	if err := putJSON(httpClient, base+"/machine-config", map[string]any{
		"vcpu_count":   1,
		"mem_size_mib": 256,
	}); err != nil {
		_ = cmd.Process.Kill()
		return nil, fmt.Errorf("failed to configure machine: %w", err)
	}

	// 3. Configure Boot Source (kernel)
	imgwd := "/var/lib/firecracker/images/"
	s.logger.InfoContext(ctx, "configuring boot source", slog.String("id", id))
	if err := putJSON(httpClient, base+"/boot-source", map[string]any{
		"kernel_image_path": filepath.Join(imgwd, "vmlinux"),
		"boot_args":         "console=ttyS0 reboot=k panic=1 pci=off",
	}); err != nil {
		_ = cmd.Process.Kill()
		return nil, fmt.Errorf("failed to configure boot source: %w", err)
	}

	// 4. Configure Root Filesystem Drive
	s.logger.InfoContext(ctx, "configuring rootfs", slog.String("id", id))
	if err := putJSON(httpClient, base+"/drives/rootfs", map[string]any{
		"drive_id":     "rootfs",
		"path_on_host": filepath.Join(imgwd, "rootfs.ext4"),
		"is_root_device": true,
		"is_read_only": false,
	}); err != nil {
		_ = cmd.Process.Kill()
		return nil, fmt.Errorf("failed to configure rootfs: %w", err)
	}

	// 5. Start Instance
	s.logger.InfoContext(ctx, "starting microVM instance", slog.String("id", id))
	if err := putJSON(httpClient, base+"/actions", map[string]string{
		"action_type": "InstanceStart",
	}); err != nil {
		_ = cmd.Process.Kill()
		return nil, fmt.Errorf("failed to start instance: %w", err)
	}

	s.logger.InfoContext(ctx, "microVM successfully booted", slog.String("id", id))

	return &pb.NodeInstance{
		Id:         id,
		State:      pb.NodeState_RUNNING,
		CreateTime: timestamppb.Now(),
	}, nil
}

func (s *InstanceServer) teardown() error {
	return nil
}

func putJSON(client *http.Client, url string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
	if err != nil {
		return err
	}

	// Firecracker expects application/json
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed to %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status %d for %s", resp.StatusCode, url)
	}

	return nil
}

