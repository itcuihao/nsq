package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/itcuihao/nsq-note/internal/lg"
	"github.com/itcuihao/nsq-note/internal/version"
	"github.com/itcuihao/nsq-note/nsqlookupd"
	"github.com/judwhite/go-svc/svc"
	"github.com/mreiferson/go-options"
)

func nsqlookupdFlagSet(opts *nsqlookupd.Options) *flag.FlagSet {
	flagSet := flag.NewFlagSet("nsqlookupd", flag.ExitOnError)

	flagSet.String("config", "", "path to config file")
	flagSet.Bool("version", false, "print version string")

	logLevel := opts.LogLevel
	flagSet.Var(&logLevel, "log-level", "set log verbosity: debug, info, warn, error, or fatal")
	flagSet.String("log-prefix", "[nsqlookupd] ", "log message prefix")
	flagSet.Bool("verbose", false, "[deprecated] has no effect, use --log-level")

	flagSet.String("tcp-address", opts.TCPAddress, "<addr>:<port> to listen on for TCP clients")
	flagSet.String("http-address", opts.HTTPAddress, "<addr>:<port> to listen on for HTTP clients")
	flagSet.String("broadcast-address", opts.BroadcastAddress, "address of this lookupd node, (default to the OS hostname)")

	flagSet.Duration("inactive-producer-timeout", opts.InactiveProducerTimeout, "duration of time a producer will remain in the active list since its last ping")
	flagSet.Duration("tombstone-lifetime", opts.TombstoneLifetime, "duration of time a producer will remain tombstoned if registration remains")

	return flagSet
}

// nsqlookupd 启动服务实例
type program struct {
	once       sync.Once
	nsqlookupd *nsqlookupd.NSQLookupd
}

func main() {
	prg := &program{}

	// svc.run方法接收一个实现了init,start和stop方法的服务实例，
	// 以及若干信号；信号用于控制该服务的优雅终止，而服务实例用于开启nsqlookupd服务
	if err := svc.Run(prg, syscall.SIGINT, syscall.SIGTERM); err != nil {
		logFatal("%s", err)
	}
}

// 实现 Init, Start, Stop 方法

func (p *program) Init(env svc.Environment) error {
	if env.IsWindowsService() {
		dir := filepath.Dir(os.Args[0])
		return os.Chdir(dir)
	}
	return nil
}

func (p *program) Start() error {
	opts := nsqlookupd.NewOptions()

	flagSet := nsqlookupdFlagSet(opts)
	flagSet.Parse(os.Args[1:])

	// version 查看版本并退出
	if flagSet.Lookup("version").Value.(flag.Getter).Get().(bool) {
		fmt.Println(version.String("nsqlookupd"))
		os.Exit(0)
	}

	// 解析config至cfg中
	var cfg map[string]interface{}
	configFile := flagSet.Lookup("config").Value.String()
	if configFile != "" {
		_, err := toml.DecodeFile(configFile, &cfg)
		if err != nil {
			logFatal("failed to load config file %s - %s", configFile, err)
		}
	}

	// 利用从用户输入的配置中的设置，替换 opts 中的参数
	options.Resolve(opts, flagSet, cfg)

	// 初始化 nsqlookupd
	nsqlookupd, err := nsqlookupd.New(opts)
	if err != nil {
		logFatal("failed to instantiate nsqlookupd", err)
	}

	// 给 p 的 nsqlookupd 赋值
	p.nsqlookupd = nsqlookupd

	go func() {
		// nsqlookupd模块启动的主体函数；
		// 当这个Start函数返回之后，整个程序阻塞在svc.Run方法内部的信号channel上；
		// 当我们向这个程序发送SIGINT和SIGTERM信号时，
		// svc.Run函数调用program.Stop方法终止nsqlookupd进程。
		err := p.nsqlookupd.Main()
		if err != nil {
			p.Stop()
			os.Exit(1)
		}
	}()

	return nil
}

func (p *program) Stop() error {
	// 利用 Sync.Once 实现类似单例模式的控制
	p.once.Do(func() {
		p.nsqlookupd.Exit()
	})
	return nil
}

func logFatal(f string, args ...interface{}) {
	lg.LogFatal("[nsqlookupd] ", f, args...)
}
