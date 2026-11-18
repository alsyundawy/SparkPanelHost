//go:build darwin

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

	var xucred *unix.Xucred
	var sockErr error
	err = sysConn.Control(func(fd uintptr) {
		xucred, sockErr = unix.GetsockoptXucred(int(fd), unix.SOL_LOCAL, unix.LOCAL_PEERCRED)
	})
	if err != nil {
		return PeerAuthInfo{}, err
	}
	if sockErr != nil {
		return PeerAuthInfo{}, sockErr
	}

	return PeerAuthInfo{UID: xucred.Uid, PID: 0}, nil
}
