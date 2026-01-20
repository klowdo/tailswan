package supervisor

import (
	"fmt"
	"log"
	"os/exec"
)

func SetupSystem() error {
	log.Println("Enabling IP forwarding...")

	sysctlParams := []struct{ key, value string }{
		{"net.ipv4.ip_forward", "1"},
		{"net.ipv6.conf.all.forwarding", "1"},
		{"net.ipv4.conf.all.send_redirects", "0"},
		{"net.ipv4.conf.default.send_redirects", "0"},
	}

	for _, p := range sysctlParams {
		cmd := exec.Command("sysctl", "-w", fmt.Sprintf("%s=%s", p.key, p.value))
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("sysctl %s: %w", p.key, err)
		}
	}

	log.Println("Setting up iptables rules...")

	iptablesRules := [][]string{
		{"iptables", "-t", "nat", "-A", "POSTROUTING", "-o", "tailscale0", "-j", "MASQUERADE"},
		{"ip6tables", "-t", "nat", "-A", "POSTROUTING", "-o", "tailscale0", "-j", "MASQUERADE"},
	}

	for _, rule := range iptablesRules {
		cmd := exec.Command(rule[0], rule[1:]...)
		_ = cmd.Run()
	}

	return nil
}
