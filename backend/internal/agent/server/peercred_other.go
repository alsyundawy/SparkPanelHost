//go:build !linux && !darwin

package server

import (
	"errors"
	"net"
)

func getPeerCred(rawConn net.Conn) (PeerAuthInfo, error) {
	if _, ok := rawConn.(*net.UnixConn); !ok {
		return PeerAuthInfo{}, errors.New("hanya koneksi unix domain socket yang diizinkan")
	}
	return PeerAuthInfo{UID: 0, PID: 0}, nil
}
