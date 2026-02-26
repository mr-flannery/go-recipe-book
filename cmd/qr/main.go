package main

import (
	"fmt"
	"net"
	"os"

	"github.com/mdp/qrterminal/v3"
)

func main() {
	ip := getLocalIP()
	if ip == "" {
		fmt.Fprintln(os.Stderr, "Could not determine local IP address")
		os.Exit(1)
	}

	port := "8080"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	url := fmt.Sprintf("http://%s:%s", ip, port)

	fmt.Printf("Scan to open: %s\n\n", url)
	qrterminal.GenerateWithConfig(url, qrterminal.Config{
		Level:     qrterminal.L,
		Writer:    os.Stdout,
		BlackChar: qrterminal.BLACK,
		WhiteChar: qrterminal.WHITE,
		QuietZone: 1,
	})
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
