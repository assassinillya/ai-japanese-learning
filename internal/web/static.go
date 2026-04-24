package web

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed assets/*
var assets embed.FS

type StaticServer struct {
	handler http.Handler
}

func NewStaticServer() *StaticServer {
	sub, err := fs.Sub(assets, "assets")
	if err != nil {
		panic(err)
	}
	return &StaticServer{
		handler: http.FileServer(http.FS(sub)),
	}
}

func (s *StaticServer) Handler() http.Handler {
	return s.handler
}
