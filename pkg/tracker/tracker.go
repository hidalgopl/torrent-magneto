package tracker

import (
	"encoding/binary"
	"errors"
	"math/rand"
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
	peerID := ""
	return &AnnounceReq{
		ConnectionID:  connectionID,
		Action:        1,
		TransactionID: rand.Uint32(),
		InfoHash:      infoHash,
		PeerID:        bytes(peerID),
		Downloaded:    0,
		Left:          0,
		Uploaded:      0,
		Event:         0,
		IPAddress:     0,
		Key:           0,
		NumWant:       -1,
		Port:          0,
	}
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
