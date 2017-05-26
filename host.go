package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

type hostEntry struct {
	hostname string
	ip       net.IP
	// Aliases are read from the /etc/hosts file, but not mapped to Route53
	aliases []string
}

type hostList []hostEntry

func (h hostList) Len() int {
	return len(h)
}

func (h hostList) Less(i, j int) bool {
	if h[i].hostname != h[j].hostname {
		return h[i].hostname < h[j].hostname
	}
	return bytes.Compare(h[i].ip, h[j].ip) < 0
}

func (h hostList) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func parseLine(line string) (*hostEntry, error) {
	if i := strings.Index(line, "#"); i >= 0 {
		line = line[0:i]
	}

	parts := strings.Fields(line)
	if len(parts) == 0 {
		return nil, nil
	}

	if len(parts) < 2 {
		return nil, fmt.Errorf("should contain at least two fields")
	}

	if ip := net.ParseIP(parts[0]); ip != nil {
		return &hostEntry{parts[1], ip, parts[2:]}, nil
	}

	return nil, fmt.Errorf("%s is not a valid IP", parts[0])
}

func readHosts(filename string) (hosts hostList) {
	file, err := os.Open(filename)

	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	i := 0
	for scanner.Scan() {
		i++
		host, err := parseLine(scanner.Text())
		if err != nil {
			log.Printf("WARN %v on line %v, skipping\n", err, i)
			continue
		}
		if host != nil {
			hosts = append(hosts, *host)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return
}

func filterHosts(hosts hostList, networks []CIDRNet) hostList {
	output := hostList{}
	for _, host := range hosts {
		for _, net := range networks {
			if net.Contains(host.ip) {
				output = append(output, host)
				break
			}
		}
	}
	return output
}
