package http_api

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/itcuihao/nsq-note/internal/lg"
)

type logWriter struct {
	logf lg.AppLogFunc
}

func (l logWriter) Write(p []byte) (int, error) {
	l.logf(lg.WARN, "%s", string(p))
	return len(p), nil
}

// 实例化http.Server模块，然后调用server.Server(listner)函数开启http服务；
func Serve(listener net.Listener, handler http.Handler, proto string, logf lg.AppLogFunc) error {
	logf(lg.INFO, "%s: listening on %s", proto, listener.Addr())

	server := &http.Server{
		Handler:  handler,
		ErrorLog: log.New(logWriter{logf}, "", 0),
	}
	err := server.Serve(listener)
	// theres no direct way to detect this error because it is not exposed
	if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
		return fmt.Errorf("http.Serve() error - %s", err)
	}

	logf(lg.INFO, "%s: closing %s", proto, listener.Addr())

	return nil
}
