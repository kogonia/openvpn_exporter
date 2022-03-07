package exporters

import (
	"bufio"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

func parseStatusFile() error {
	f, err := os.Open(ovpn.statusPath)
	if err != nil {
		return err
	}
	defer f.Close()

	ovpn.clients = make(map[string]client, 128)

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		processRow(scanner.Text())
	}
	// countTotalBytes()

	return nil
}

func processRow(row string) {
	fields := strings.Split(row, ",")
	switch len(fields) {
	case 2:
		if fields[0] == "updated" {
			ovpn.updated = fields[1]
		}

	case 5:
		if fields[0] == "Common Name" {
			// client list header. Pass through
		} else {
			// client list entry
			err := processClientListEntry(fields)
			if err != nil {
				log.Println(err)
			}
		}

	case 4:
		if fields[0] == "Virtual Address" {
			// routing table header. Pass through
		} else {
			// routing table entry
			processRoutingTableEntry(fields)
		}
	}
}

func processClientListEntry(fields []string) error {
	br, err := strconv.ParseFloat(fields[2], 64)
	if err != nil {
		log.Println(err)
		return err
	}
	bs, err := strconv.ParseFloat(fields[3], 64)
	if err != nil {
		log.Println(err)
		return err
	}

	ovpn.totalBytesIn += br
	ovpn.totalBytesOut += bs

	// t, _ := time.Parse("2006-01-02 15:04:05", fields[4])
	cl := client{
		commonName:    fields[0],
		realAddr:      net.ParseIP(strings.Split(fields[1], ":")[0]),
		connSince:     fields[4],
		bytesReceived: br,
		bytesSent:     bs,
	}
	ovpn.clients[cl.commonName] = cl

	return nil
}

func processRoutingTableEntry(fields []string) {
	cn := fields[1]
	if cl, ok := ovpn.clients[cn]; ok {
		// t, _ := time.Parse("2006-01-02 15:04:05", fields[3])
		cl.virtAddr = net.ParseIP(strings.Split(fields[0], ":")[0])
		cl.lastRef = fields[3]
		ovpn.clients[cn] = cl
	}
}
