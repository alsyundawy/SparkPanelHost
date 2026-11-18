package server

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc/credentials"
)

type PeerAuthInfo struct {
	UID uint32
	PID int32
}

func (PeerAuthInfo) AuthType() string { return "unix-peercred" }

type PeerCreds struct {
	AllowedUID uint32
}

func (p *PeerCreds) ServerHandshake(rawConn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	info, err := getPeerCred(rawConn)
	if err != nil {
		return nil, nil, err
	}

	if info.UID != p.AllowedUID && info.UID != 0 {
		return nil, nil, fmt.Errorf("security violation: akses socket ditolak untuk UID %d (PID %d)", info.UID, info.PID)
	}

	return rawConn, info, nil
}

func (p *PeerCreds) ClientHandshake(_ context.Context, _ string, rawConn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	return rawConn, nil, nil
}

func (p *PeerCreds) Info() credentials.ProtocolInfo {
	return credentials.ProtocolInfo{SecurityProtocol: "unix-peercred"}
}

func (p *PeerCreds) Clone() credentials.TransportCredentials {
	return &PeerCreds{AllowedUID: p.AllowedUID}
}

func (p *PeerCreds) OverrideServerName(_ string) error { return nil }
