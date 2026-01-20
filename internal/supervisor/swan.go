package supervisor

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
)

type SwanService struct{}

func (sw *SwanService) LoadConfig(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("config not found: %s", path)
	}

	slog.Info("Loading swanctl configuration", "path", path)

	cmd := exec.Command("swanctl", "--load-all")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (sw *SwanService) Initiate(connection string) error {
	slog.Info("Initiating connection: %s", connection)

	cmd := exec.Command("swanctl", "--initiate", "--child", connection)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("initiate %s: %w\nOutput: %s", connection, err, output)
	}

	return nil
}

func (sw *SwanService) Terminate(connection string) error {
	slog.Info("Terminating connection: %s", connection)

	cmd := exec.Command("swanctl", "--terminate", "--ike", connection)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("terminate %s: %w\nOutput: %s", connection, err, output)
	}

	return nil
}

func (sw *SwanService) ListConnections() error {
	cmd := exec.Command("swanctl", "--list-conns")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (sw *SwanService) ListSAs() error {
	cmd := exec.Command("swanctl", "--list-sas")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (sw *SwanService) Reload() error {
	slog.Info("Reloading swanctl configuration...")
	cmd := exec.Command("swanctl", "--load-all")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
