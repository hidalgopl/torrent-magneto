package main

import (
	"fmt"
	"github.com/hidalgopl/torrent-magneto/pkg/tracker"
	"net"
	"os"
	"strings"
	//"strings"
)


func main() {
	// TODO - read from stdin
	//reader := bufio.NewReader(os.Stdin)
	//fmt.Print("Paste magnet link>> ")
	//magnetLink, _ := reader.ReadString('\n')
	magnetLink := "magnet:?xt=urn:btih:7FBC58E324B539BDDA58C15BDA3ACD26B0D5FBD1&dn=Luis%20Fonsi%20-%20Despacito%20(feat.%20Daddy%20Yankee)&tr=udp%3A%2F%2Ftracker.coppersurfer.tk%3A6969%2Fannounce&tr=udp%3A%2F%2F9.rarbg.to%3A2920%2Fannounce&tr=udp%3A%2F%2Ftracker.opentrackr.org%3A1337&tr=udp%3A%2F%2Ftracker.internetwarriors.net%3A1337%2Fannounce&tr=udp%3A%2F%2Ftracker.coppersurfer.tk%3A6969%2Fannounce&tr=udp%3A%2F%2Ftracker.pirateparty.gr%3A6969%2Fannounce&tr=udp%3A%2F%2Ftracker.cyberia.is%3A6969%2Fannounce"
	pm, err := tracker.ParseMagnetLink(magnetLink)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, tr := range pm.Tr {
		splitAnnounce := strings.TrimSuffix(tr, "/announce")
		fmt.Println(splitAnnounce)
		splitProto := strings.TrimPrefix(splitAnnounce, "udp://")
		fmt.Println(splitProto)
		s, err := net.ResolveUDPAddr("udp", splitProto)
		if err != nil {
			fmt.Println(err)
			return
		}
		c, err := net.DialUDP("udp", nil, s)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("The UDP server is %s\n", c.RemoteAddr().String())
		defer c.Close()

		b, err := tracker.NewConnectReq()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		req, err := b.Serialize()
		_, err = c.Write(req)
		fmt.Println("wrote:")
		fmt.Println(b)

		if err != nil {
			fmt.Println("error:")
			panic(err)
			os.Exit(1)
			//return
		}
		fmt.Println("attempting to read")
		buffer := make([]byte, 16)
		n, _, err := c.ReadFromUDP(buffer)

		fmt.Println("waiting for reply")
		if err != nil {
			fmt.Println(err)
		}
		connectRsp := tracker.SerializeConnectResp(buffer[0:n])
		fmt.Printf("action: %v, transactionID: %v, connectionID: %v \n",
			connectRsp.Action, connectRsp.TransactionID, connectRsp.ConnectionID)

		scrapeReq := tracker.NewScrapeReq(pm.GetInfoHash(), connectRsp.ConnectionID)
		req, err = scrapeReq.Serialize()
		_, err = c.Write(req)
		if err != nil {
			fmt.Println(err)
		}
		bufferScrape := make([]byte, 100000)
		n, _, err = c.ReadFromUDP(bufferScrape)
		respBody, err := tracker.DeserializeScrapeResp(bufferScrape)
		fmt.Printf("Scrape reply: %v \n", respBody)
	}


}
