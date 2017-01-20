package api

import (
	"crypto/tls"
	"net"
	"os"
	"syscall"
)

func newListener(proto string, addr string, tlsConfig *tls.Config) (net.Listener, error) {

	l, err := net.Listen(proto, addr)
	if err != nil {
		return nil, err
	}

	if tlsConfig != nil {
		tlsConfig.NextProtos = []string{"http/1.1"}
		l = tls.NewListener(l, tlsConfig)
	}
	return l, nil
}

func newUnixListener(addr string, tlsConfig *tls.Config) (net.Listener, error) {

	if err := syscall.Unlink(addr); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	mask := syscall.Umask(0777)
	defer syscall.Umask(mask)
	l, err := newListener("unix", addr, tlsConfig)
	if err != nil {
		return nil, err
	}

	if err := os.Chmod(addr, 0600); err != nil {
		return nil, err
	}
	return l, nil
}
