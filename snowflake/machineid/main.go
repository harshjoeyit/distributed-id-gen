package machineid

import (
	"hash/fnv"
	"log"
	"net"
)

// getMachineID returns the machine ID (10 bits) of the current machine
func Get() (int, error) {
	// return getMachineIDFromCentralService()
	// return 1, nil
	return hashIPtoMachineID()
}

// Hashes an IP address to generate a machine ID (10 bits)
func hashIPtoMachineID() (int, error) {
	// find private IP address of the machine
	ip, err := getPrivateIP()
	if err != nil {
		return 0, err
	}

	// fast, non-cryptographic hash
	h := fnv.New32a()
	h.Write([]byte(ip))
	mID := int(h.Sum32() % 1024) // 10 bit ID

	return mID, nil
}

// isPrivateIP checks if an IP address is private
func isPrivateIP(ip net.IP) bool {
	privateIPv4Ranges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
	}

	for _, cidr := range privateIPv4Ranges {
		_, ipNet, _ := net.ParseCIDR(cidr)

		if ipNet.Contains(ip) {
			return true
		}
	}

	privateIPv6Range := "fc00::/7"
	_, ipNet, _ := net.ParseCIDR(privateIPv6Range)
	return ipNet.Contains(ip)
}

func getPrivateIP() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Printf("failed to get network interfaces: %v", err)
	}

	var privateIPs []string

	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}

			if isPrivateIP(ip) {
				privateIPs = append(privateIPs, ip.String())
			}
		}
	}

	if len(privateIPs) == 0 {
		return "", nil
	}

	return privateIPs[0], nil
}
