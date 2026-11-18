//go:build linux

package server

import (
	"errors"
	"net"

	"golang.org/x/sys/unix"
)

func getPeerCred(rawConn net.Conn) (PeerAuthInfo, error) {
	unixConn, ok := rawConn.(*net.UnixConn)
	if !ok {
		return PeerAuthInfo{}, errors.New("hanya koneksi unix domain socket yang diizinkan")
	}

	sysConn, err := unixConn.SyscallConn()
	if err != nil {
		return PeerAuthInfo{}, err
	}

	var ucred *unix.Ucred
	var sockErr error
	err = sysConn.Control(func(fd uintptr) {
		ucred, sockErr = unix.GetsockoptUcred(int(fd), unix.SOL_SOCKET, unix.SO_PEERCRED)
	})
	if err != nil {
		return PeerAuthInfo{}, err
	}
	if sockErr != nil {
		return PeerAuthInfo{}, sockErr
	}

	return PeerAuthInfo{UID: ucred.Uid, PID: ucred.Pid}, nil
}
