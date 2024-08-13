package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/miekg/dns"
)

func handleDNSRequest(w http.ResponseWriter, r *http.Request) {
	var dnsMessage []byte
	var err error

	switch r.Method {
	case http.MethodGet:
		// For GET, the DNS message is in the 'dns' query parameter
		dnsParam := r.URL.Query().Get("dns")
		if dnsParam == "" {
			http.Error(w, "Missing 'dns' query parameter", http.StatusBadRequest)
			return
		}
		dnsMessage, err = base64.RawURLEncoding.DecodeString(dnsParam)
		if err != nil {
			http.Error(w, "Failed to decode DNS message", http.StatusBadRequest)
			return
		}

	case http.MethodPost:
		// For POST, the DNS message is in the request body
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		dnsMessage = body

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the DNS message
	msg := new(dns.Msg)
	if err := msg.Unpack(dnsMessage); err != nil {
		log.Println("failure unpacking DNS message:", err)
		http.Error(w, "Failed to unpack DNS message", http.StatusBadRequest)
		return
	}

	slog.Info("request received", "msg", msg)
	// Create a DNS client
	client := &dns.Client{Net: "udp"}

	//Send the query to an upstream DNS server (Google's public DNS in this case)
	dnsServer := os.Getenv("DNS_SERVER")
	if dnsServer == "" {
		dnsServer = "203.201.60.12"
	}
	//dnsServer, err := getUnixDNSServer()
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}

	dnsServer = fmt.Sprintf("%s:%d", dnsServer, 53)

	slog.Info("requesting dns server", "server addr", dnsServer)
	response, _, err := client.Exchange(msg, dnsServer)
	if err != nil {
		log.Println("failure querying upstream DNS server:", dnsServer, err)
		http.Error(w, "Failed to resolve DNS query", http.StatusInternalServerError)
		return
	}

	// Set the response bit
	response.Response = true

	// Pack the response
	packedResponse, err := response.Pack()
	if err != nil {
		log.Println("failure packing response", err)
		http.Error(w, "Failed to pack DNS response", http.StatusInternalServerError)
		return
	}

	slog.Info("responding with msg", "resp", response)
	// Set the content type and write the response
	w.Header().Set("Content-Type", "application/dns-message")
	w.Write(packedResponse)
}

func main() {
	http.HandleFunc("/dns-query", handleDNSRequest)
	fmt.Println("Starting DNS over HTTP server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func getUnixDNSServer() (string, error) {
	file, err := os.Open("/etc/resolv.conf")
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "nameserver") {
			fields := strings.Fields(line)
			if len(fields) > 1 {
				return fields[1], nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("default DNS server not found")
}
