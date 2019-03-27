package main

import (
	"crypto/tls"
	"os"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/mreiferson/go-options"
	"github.com/itcuihao/nsq-note/internal/test"
	"github.com/itcuihao/nsq-note/nsqd"
)

func TestConfigFlagParsing(t *testing.T) {
	opts := nsqd.NewOptions()
	opts.Logger = test.NewTestLogger(t)

	flagSet := nsqdFlagSet(opts)
	flagSet.Parse([]string{})

	var cfg config
	f, err := os.Open("../../contrib/nsqd.cfg.example")
	if err != nil {
		t.Fatalf("%s", err)
	}
	toml.DecodeReader(f, &cfg)
	cfg.Validate()

	options.Resolve(opts, flagSet, cfg)
	nsqd.New(opts)

	if opts.TLSMinVersion != tls.VersionTLS10 {
		t.Errorf("min %#v not expected %#v", opts.TLSMinVersion, tls.VersionTLS10)
	}
}
