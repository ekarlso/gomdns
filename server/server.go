package server

import (
	log "code.google.com/p/log4go"
	"github.com/ekarlso/gomdns/api"
	"github.com/ekarlso/gomdns/config"
	"github.com/ekarlso/gomdns/nameserver"
)

type Server struct {
	ApiServer  *api.HttpServer
	NameServer *nameserver.NameServer
	Config     *config.Configuration
	stopped    bool
}

func NewServer(cfg *config.Configuration) (*Server, error) {
	apiServer := api.NewServer(cfg)
	nameServer := nameserver.NewServer(cfg)

	return &Server{
		ApiServer:  apiServer,
		NameServer: nameServer,
		Config:     cfg,
	}, nil
}

func (self *Server) ListenAndServe() (err error) {
	log.Debug("Starting API on %s", self.Config.ApiServerListen())
	go self.ApiServer.ListenAndServe()

	self.NameServer.ListenAndServe()

	return err
}

func (self *Server) Stop() {
	if self.stopped {
		return
	}

	self.stopped = true
}
