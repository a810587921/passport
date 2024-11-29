package forward

import (
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/yosebyte/passport/pkg/log"
)

func HandleUDP(parsedURL *url.URL, whiteList *sync.Map) error {
	linkAddr, err := net.ResolveUDPAddr("udp", parsedURL.Host)
	if err != nil {
		log.Error("Unable to resolve link address: %v", parsedURL.Host)
		return err
	}
	targetAddr, err := net.ResolveUDPAddr("udp", strings.TrimPrefix(parsedURL.Path, "/"))
	if err != nil {
		log.Error("Unable to resolve target address: %v", strings.TrimPrefix(parsedURL.Path, "/"))
		return err
	}
	linkConn, err := net.ListenUDP("udp", linkAddr)
	if err != nil {
		log.Error("Unable to listen link address: [%v]", linkAddr)
		return err
	}
	defer linkConn.Close()
	targetConn, err := net.DialUDP("udp", nil, targetAddr)
	if err != nil {
		log.Error("Unable to dial target address: [%v]", targetAddr)
		return err
	}
	defer targetConn.Close()
	log.Info("Target connection established: [%v]", targetAddr)
	buffer := make([]byte, 4096)
	for {
		n, clientAddr, err := linkConn.ReadFromUDP(buffer)
		if err != nil {
			log.Error("Unable to read from client address: [%v] %v", clientAddr, err)
			continue
		}
		if parsedURL.Fragment != "" {
			clientIP := clientAddr.IP.String()
			if _, exists := whiteList.Load(clientIP); !exists {
				log.Warn("Unauthorized IP address blocked: [%v]", clientIP)
				continue
			}
		}
		err = targetConn.SetDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			log.Error("Unable to set deadline: %v", err)
			continue
		}
		log.Info("Starting data transfer: [%v] <-> [%v]", clientAddr, targetAddr)
		_, err = targetConn.Write(buffer[:n])
		if err != nil {
			log.Error("Unable to write to target address: [%v] %v", targetAddr, err)
			continue
		}
		n, _, err = targetConn.ReadFromUDP(buffer)
		if err != nil {
			log.Error("Unable to read from target address: [%v] %v", targetAddr, err)
			continue
		}
		_, err = linkConn.WriteToUDP(buffer[:n], clientAddr)
		if err != nil {
			log.Error("Unable to write to client address: [%v] %v", clientAddr, err)
			continue
		}
		log.Info("Transfer completed successfully")

	}
}
