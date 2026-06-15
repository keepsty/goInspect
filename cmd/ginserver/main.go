// Copyright 2026

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/hanchuanchuan/goInception/server"
	log "github.com/sirupsen/logrus"
)

var (
	addr = flag.String("addr", ":8080", "gin HTTP listen address")
	mode = flag.String("mode", "release", "gin mode: debug/release/test")
)

func main() {
	flag.Parse()
	gin.SetMode(*mode)
	router := server.NewGinEngine()

	log.Infof("Starting gin server at %s", *addr)
	if err := router.Run(*addr); err != nil {
		fmt.Fprintf(os.Stderr, "failed to start gin server: %v\n", err)
		os.Exit(1)
	}
}
