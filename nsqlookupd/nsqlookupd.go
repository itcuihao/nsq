package nsqlookupd

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"github.com/itcuihao/nsq-note/internal/http_api"
	"github.com/itcuihao/nsq-note/internal/protocol"
	"github.com/itcuihao/nsq-note/internal/util"
	"github.com/itcuihao/nsq-note/internal/version"
)

type NSQLookupd struct {
	sync.RWMutex
	opts         *Options
	tcpListener  net.Listener
	httpListener net.Listener
	waitGroup    util.WaitGroupWrapper
	DB           *RegistrationDB
}

func New(opts *Options) (*NSQLookupd, error) {
	var err error

	if opts.Logger == nil {
		opts.Logger = log.New(os.Stderr, opts.LogPrefix, log.Ldate|log.Ltime|log.Lmicroseconds)
	}
	l := &NSQLookupd{
		opts: opts,
		DB:   NewRegistrationDB(),
	}

	l.logf(LOG_INFO, version.String("nsqlookupd"))

	l.tcpListener, err = net.Listen("tcp", opts.TCPAddress)
	if err != nil {
		return nil, fmt.Errorf("listen (%s) failed - %s", opts.TCPAddress, err)
	}
	l.httpListener, err = net.Listen("tcp", opts.HTTPAddress)
	if err != nil {
		return nil, fmt.Errorf("listen (%s) failed - %s", opts.TCPAddress, err)
	}

	return l, nil
}

// Main starts an instance of nsqlookupd and returns an
// error if there was a problem starting up.
func (l *NSQLookupd) Main() error {

	// 整个 nsqlookupd 的上下文
	ctx := &Context{l}

	exitCh := make(chan error)
	var once sync.Once
	exitFunc := func(err error) {
		once.Do(func() {
			if err != nil {
				l.logf(LOG_FATAL, "%s", err)
			}
			exitCh <- err
		})
	}

	// TCP 服务
	// 执行流程：
	// 1.首先在Main函数开启一个goroutine来开启一个tcp循环，接收nsqd连接请求；
	// 2.当接收到一个nsqd接请求时，开启一个goroutine来处理这个nsqd；
	// 3.这个nsqd先是经历tcpServer.Handle函数，然后到LookupProtocolV1.IOLoop函数，并阻塞在此函数中；
	// 4.当nsqd发送请求时，LookupProtocolV1.IOLoop函数先是读取该请求，并调用LookupProtocolV1.Exec函数执行具体请求；
	tcpServer := &tcpServer{ctx: ctx}

	// Wrap方法用于在新的goroutine调用参数func函数；因此在执行Main方法之后，
	// 此时nsqlookupd进程就另外开启了两个goroutine，一个用于执行tcp服务，一个用于执行http服务；

	// 开启 TCP 服务的 groutine
	l.waitGroup.Wrap(func() {
		exitFunc(protocol.TCPServer(l.tcpListener, tcpServer, l.logf))
	})

	// HTTP 服务
	// 用于向nsqadmin提供查询接口的，本质上，就是一个web服务器，提供http查询接口
	httpServer := newHTTPServer(ctx)

	// 开启 HTTP 服务的 groutine
	l.waitGroup.Wrap(func() {
		exitFunc(http_api.Serve(l.httpListener, httpServer, "HTTP", l.logf))
	})

	err := <-exitCh
	return err
}

func (l *NSQLookupd) RealTCPAddr() *net.TCPAddr {
	return l.tcpListener.Addr().(*net.TCPAddr)
}

func (l *NSQLookupd) RealHTTPAddr() *net.TCPAddr {
	return l.httpListener.Addr().(*net.TCPAddr)
}

func (l *NSQLookupd) Exit() {
	if l.tcpListener != nil {
		l.tcpListener.Close()
	}

	if l.httpListener != nil {
		l.httpListener.Close()
	}
	l.waitGroup.Wait()
}
