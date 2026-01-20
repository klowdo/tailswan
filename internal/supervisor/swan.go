package supervisor

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

type SwanService struct{}

func (sw *SwanService) LoadConfig(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("config not found: %s", path)
	}

	log.Printf("Loading swanctl configuration from %s...", path)

	cmd := exec.Command("swanctl", "--load-all")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (sw *SwanService) Initiate(connection string) error {
	log.Printf("Initiating connection: %s", connection)

	cmd := exec.Command("swanctl", "--initiate", "--child", connection)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("initiate %s: %w\nOutput: %s", connection, err, output)
	}

	return nil
}

func (sw *SwanService) Terminate(connection string) error {
	log.Printf("Terminating connection: %s", connection)

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
	log.Println("Reloading swanctl configuration...")
	cmd := exec.Command("swanctl", "--load-all")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
