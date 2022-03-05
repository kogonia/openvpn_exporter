package exporters

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type openVPN struct {
	updated                     string
	totalBytesIn, totalBytesOut int
	clients                     map[string]client
}

type client struct {
	comonName                string
	virtAddr, realAddr       net.IP
	connSince, lastRef       time.Time
	bytesReceived, bytesSent int
}

var ov openVPN

func (ov openVPN) bytes() []byte {
	b := bytes.NewBuffer(make([]byte, 0, 64))
	_ = json.NewEncoder(b).Encode(ov)
	return b.Bytes()
}

func Process(filename string) {
	ov.clients = make(map[string]client, 128)
	parseStatusFile(filename)
}

func printer(ov openVPN) {
	fmt.Printf("updated: %s\n", ov.updated)
	fmt.Printf("totalBytesIn: %d\n", ov.totalBytesIn)
	fmt.Printf("totalBytesOut: %d\n[\n", ov.totalBytesOut)
	for _, cl := range ov.clients {
		fmt.Printf("{\n\tcomonName:\t%s,\n\tvirtAddr:\t%s,\n\trealAddr:\t%s,\n\tconnSince:\t%s,\n\tlastRef:\t%s,\n\tbytesReceived:\t%d,\n\tbytesSent:\t%d\n},\n",
			cl.comonName, cl.virtAddr, cl.realAddr, cl.connSince, cl.lastRef, cl.bytesReceived, cl.bytesSent)
	}
	fmt.Println("]")
}

func check(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func parseStatusFile(fileName string) {
	f, err := os.Open(fileName)
	check(err)
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		processRow(scanner.Text())
	}
	countTotalBytes()
	printer(ov)
}

func processRow(row string) {
	fields := strings.Split(row, ",")
	switch len(fields) {
	case 2:
		if fields[0] == "Updated" {
			ov.updated = fields[1]
		}

	case 5:
		if fields[0] == "Common Name" {
			// client list header. Pass through
		} else {
			processClientListEntry(fields)
		} // client list entry

	case 4:
		if fields[0] == "Virtual Address" {
			// routing table header. Pass through
		} else {
			// routing table entry
			processRoutingTableEntry(fields)
		}
	}
}

func processClientListEntry(fields []string) {
	br, err := strconv.Atoi(fields[2])
	check(err)
	bs, err := strconv.Atoi(fields[3])
	check(err)

	t, _ := time.Parse("2006-01-02 15:04:05", fields[4])
	cl := client{
		comonName:     fields[0],
		realAddr:      net.ParseIP(strings.Split(fields[1], ":")[0]),
		connSince:     t,
		bytesReceived: br,
		bytesSent:     bs,
	}
	ov.clients[cl.comonName] = cl
}

func processRoutingTableEntry(fields []string) {
	cn := fields[1]
	if cl, ok := ov.clients[cn]; ok {
		t, _ := time.Parse("2006-01-02 15:04:05", fields[3])
		cl.virtAddr = net.ParseIP(strings.Split(fields[0], ":")[0])
		cl.lastRef = t
		ov.clients[cn] = cl
	}
}

func countTotalBytes() {
	var in, out int
	for _, cl := range ov.clients {
		in += cl.bytesReceived
		out += cl.bytesSent
	}
	ov.totalBytesIn = in
	ov.totalBytesOut = out
}
