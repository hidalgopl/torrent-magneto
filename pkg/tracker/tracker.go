package tracker

import (
	"encoding/binary"
	"math/rand"
	"net/url"
	"strings"
)

type ConnectReq struct {
	transactionID uint32
	action        uint32
	protocolID    uint64
}

type ConnectResp struct {
	action        uint32
	transactionID uint32
	connectionID  uint64
}

type ScrapeReq struct {
	connectionID  uint64
	action        uint32
	transactionID uint32
	infoHash      [20]byte
}

func NewScrapeReq (infoHash [20]byte, connectionID uint64, transactionID uint32) *ScrapeReq {
	return &ScrapeReq{
		connectionID:  connectionID,
		action:        2,
		transactionID: transactionID,
		infoHash:      infoHash,
	}
}


func (sr *ScrapeReq) Serialize() []byte {
	buf := make([]byte, 36)
	binary.BigEndian.PutUint64(buf[0:8], sr.connectionID)
	binary.BigEndian.PutUint32(buf[8:12], sr.action)
	binary.BigEndian.PutUint32(buf[12:16], sr.transactionID)
	copy(buf[16:], sr.infoHash[:])
	return buf
}

func SerializeConnectResp(resp []byte) *ConnectResp {
	return &ConnectResp{
		action:        binary.BigEndian.Uint32(resp[0:4]),
		transactionID: binary.BigEndian.Uint32(resp[4:8]),
		connectionID:  binary.BigEndian.Uint64(resp[8:16]),
	}
}

func newConnectReq() (*ConnectReq, error) {
	transactionID := rand.Uint32()
	var action uint32 = 0
	var protocolID uint64 = 0x41727101980
	return &ConnectReq{
		transactionID: transactionID,
		action:        action,
		protocolID:    protocolID,
	}, nil
}

type Urn struct {
	btih Xt
}
type Xt struct {
	btih string
}

type ParsedMagnet struct {
	Dn []string
	Tr []string
	Xt []string
}

func (pm *ParsedMagnet) getInfoHash() [20]byte {
	urn := strings.Split(pm.Xt[0], ":")[2]
	urnByte := [20]byte{}
	copy(urnByte[:], urn)
	return urnByte
}

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

func (c *ConnectReq) Serialize() []byte {
	buf := make([]byte, 16)
	binary.BigEndian.PutUint64(buf[0:8], c.protocolID)
	binary.BigEndian.PutUint32(buf[8:12], c.action)
	binary.BigEndian.PutUint32(buf[12:16], c.transactionID)
	return buf
}
