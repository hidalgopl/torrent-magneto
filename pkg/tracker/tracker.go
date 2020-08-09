package tracker

import (
	"encoding/binary"
	"errors"
	"math/rand"
	"net"
	"net/url"
	"strings"
)

const (
	ActionConnect  = 0
	ActionAnnounce = 1
	ActionScrape   = 2
	ActionError    = 3
)

type Request interface {
	Serialize() ([]byte, error)
}

type Response interface {
	Deserialize() (interface{}, error)
}

// ...
type ConnectReq struct {
	TransactionID uint32
	Action        uint32
	ProtocolID    uint64
}

type ConnectResp struct {
	Action        uint32
	TransactionID uint32
	ConnectionID  uint64
}

type AnnounceReq struct {
	ConnectionID  uint64
	Action        uint32
	TransactionID uint32
	InfoHash      [20]byte
	PeerID        [20]byte
	Downloaded    uint64
	Left          uint64
	Uploaded      uint64
	Event         uint32
	IPAddress     uint32
	Key           uint32
	NumWant       uint32
	Port          uint16
}

type AnnounceResp struct {
	Action        uint32
	TransactionID uint32
	Interval      uint32
	Leechers      uint32
	Seeders       uint32
	IPAddress     uint32
	TCPPort       uint16
}

type ScrapeReq struct {
	connectionID  uint64
	action        uint32
	transactionID uint32
	infoHash      [20]byte
}

type ScrapeResp struct {
	Action        uint32
	TransactionID uint32
	Seeders       uint32
	Completed     uint32
	leechers      uint32
}

type ScrapeRespErr struct {
	action        uint32
	transactionID uint32
	message       string
}

func NewAnnounceReq(infoHash [20]byte, connectionID uint64) *AnnounceReq {
	peerID := "-BJN-s3xfj3ksloweisj"
	var peerIDbytes [20]byte
	copy(peerIDbytes[:], peerID)
	return &AnnounceReq{
		ConnectionID:  connectionID,
		Action:        1,
		TransactionID: rand.Uint32(),
		InfoHash:      infoHash,
		PeerID:        peerIDbytes,
		Downloaded:    0,
		Left:          0,
		Uploaded:      0,
		Event:         0,
		IPAddress:     0,
		Key:           0,
		NumWant:       1,
		Port:          0,
	}
}

func (aq *AnnounceReq) Serialize() ([]byte, error) {
	buf := make([]byte, 98)
	binary.BigEndian.PutUint64(buf[0:8], aq.ConnectionID)
	binary.BigEndian.PutUint32(buf[8:12], aq.Action)
	binary.BigEndian.PutUint32(buf[12:16], aq.TransactionID)
	copy(buf[16:36], aq.InfoHash[:])
	copy(buf[36:56], aq.PeerID[:])
	binary.BigEndian.PutUint64(buf[56:64], aq.Downloaded)
	binary.BigEndian.PutUint64(buf[64:72], aq.Left)
	binary.BigEndian.PutUint64(buf[72:80], aq.Uploaded)
	binary.BigEndian.PutUint32(buf[80:84], aq.Event)
	binary.BigEndian.PutUint32(buf[84:88], aq.IPAddress)
	binary.BigEndian.PutUint32(buf[88:92], aq.Key)
	binary.BigEndian.PutUint32(buf[92:96], aq.NumWant)
	binary.BigEndian.PutUint16(buf[96:98], aq.Port)

	return buf, nil
}

func DeserializeAnnounceResp(resp []byte) (*AnnounceResp, error) {
	respBody := &AnnounceResp{
		Action:        binary.BigEndian.Uint32(resp[0:4]),
		TransactionID: binary.BigEndian.Uint32(resp[4:8]),
		Interval:      binary.BigEndian.Uint32(resp[8:12]),
		Leechers:      binary.BigEndian.Uint32(resp[12:16]),
		Seeders:       binary.BigEndian.Uint32(resp[16:20]),
		IPAddress:     binary.BigEndian.Uint32(resp[20:24]),
		TCPPort:       binary.BigEndian.Uint16(resp[24:26]),
	}
	return respBody, nil
}

func (ar *AnnounceResp)GetIPAddress() net.IP {
	addr := make(net.IP, 4)
	binary.BigEndian.PutUint32(addr, ar.IPAddress)
	return addr
}

func DeserializeScrapeResp(resp []byte) (interface{}, error) {
	if len(resp) < 8 {
		errMsg := "received too small packet " + string(len(resp))
		return nil, errors.New(errMsg)
	}
	action := binary.BigEndian.Uint32(resp[0:4])
	if action == ActionError {
		return DeserializeScrapeRespErr(resp)
	}
	respBody := &ScrapeResp{
		Action:        binary.BigEndian.Uint32(resp[0:4]),
		TransactionID: binary.BigEndian.Uint32(resp[4:8]),
		Seeders:       binary.BigEndian.Uint32(resp[8:12]),
		Completed:     binary.BigEndian.Uint32(resp[12:16]),
		leechers:      binary.BigEndian.Uint32(resp[16:20]),
	}
	return respBody, nil
}

func DeserializeScrapeRespErr(resp []byte) (*ScrapeRespErr, error) {
	respBody := &ScrapeRespErr{
		action:        binary.BigEndian.Uint32(resp[0:4]),
		transactionID: binary.BigEndian.Uint32(resp[4:8]),
		message:       string(resp[8:]),
	}
	return respBody, nil
}

func NewScrapeReq(infoHash [20]byte, connectionID uint64) *ScrapeReq {
	return &ScrapeReq{
		connectionID:  connectionID,
		action:        2,
		transactionID: rand.Uint32(),
		infoHash:      infoHash,
	}
}

func (sr *ScrapeReq) Serialize() ([]byte, error) {
	buf := make([]byte, 36)
	binary.BigEndian.PutUint64(buf[0:8], sr.connectionID)
	binary.BigEndian.PutUint32(buf[8:12], sr.action)
	binary.BigEndian.PutUint32(buf[12:16], sr.transactionID)
	copy(buf[16:], sr.infoHash[:])
	return buf, nil
}

func SerializeConnectResp(resp []byte) *ConnectResp {
	return &ConnectResp{
		Action:        binary.BigEndian.Uint32(resp[0:4]),
		TransactionID: binary.BigEndian.Uint32(resp[4:8]),
		ConnectionID:  binary.BigEndian.Uint64(resp[8:16]),
	}
}

func NewConnectReq() (*ConnectReq, error) {
	transactionID := rand.Uint32()
	var action uint32 = 0
	var protocolID uint64 = 0x41727101980
	return &ConnectReq{
		TransactionID: transactionID,
		Action:        action,
		ProtocolID:    protocolID,
	}, nil
}

type ParsedMagnet struct {
	Dn []string
	Tr []string
	Xt []string
}

func (pm *ParsedMagnet) GetInfoHash() [20]byte {
	urn := strings.Split(pm.Xt[0], ":")[2]
	urnByte := [20]byte{}
	copy(urnByte[:], urn)
	return urnByte
}

// ParseMagnetLink ...
func ParseMagnetLink(magnetLink string) (*ParsedMagnet, error) {
	u, err := url.Parse(magnetLink)
	if err != nil {
		return nil, err
	}
	params, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return nil, err
	}
	return &ParsedMagnet{
		Xt: params["xt"],
		Tr: params["tr"],
		Dn: params["dn"],
	}, nil
}

func (c *ConnectReq) Serialize() ([]byte, error) {
	buf := make([]byte, 16)
	binary.BigEndian.PutUint64(buf[0:8], c.ProtocolID)
	binary.BigEndian.PutUint32(buf[8:12], c.Action)
	binary.BigEndian.PutUint32(buf[12:16], c.TransactionID)
	return buf, nil
}
