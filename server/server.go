package server

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	// "golang.org/x/time/rate"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// Implements a server with basic authentication and request rate
// Limiting. Only supports websockets connection
type BasicServer struct {
	// Base options inherited from task Executor
	Entrypoint string
}

func (b *BasicServer) Start(useTLS bool) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", b.serveWS)

	srv := &http.Server{
		Addr:         "localhost:9090",
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	log.Printf("starting websockets server on %s", srv.Addr)
	// Use FiloSottile's [mkcert](https://github.com/FiloSottile/mkcert) utility
	var err error
	if useTLS {
		err = srv.ListenAndServeTLS("localhost.pem", "localhost-key.pem")
	} else {
		err = srv.ListenAndServe()
	}
	if err != nil {
		log.Fatal("server error: ", err)
		return err
	}
	return nil
}

func (b *BasicServer) startExecution(ctx context.Context, c *websocket.Conn) error {
	r := TaskReq{}
	if err := wsjson.Read(ctx, c, &r); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}

	if err := ParseAndRun(ctx, c, r, b); err != nil {
		return err
	}

	return nil
}

// Listens for ws:// or wss:// connections
// Example modified from the official documentation for nhooyr.io/websocket
// https://github.com/nhooyr/websocket/blob/v1.8.7/examples/echo/server.go
func (b *BasicServer) serveWS(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{})
	if err != nil {
		log.Fatalf("failed to upgrade connection to websockets: %v", err)
		return
	}
	defer c.Close(websocket.StatusInternalError, "internal error")

	// l := rate.NewLimiter(rate.Every(time.Millisecond*100), 10)
	log.Printf("new connection %v\n", r.RemoteAddr)
	for {
		err = b.startExecution(r.Context(), c)

		switch websocket.CloseStatus(err) {
		case websocket.StatusNormalClosure:
			return
		case statusTaskSuccess:
			c.Close(statusTaskSuccess, "task exited successfully")
			return
		case statusTaskFailure:
			c.Close(statusTaskFailure, "running task failed")
			// Do not exit before the error is printed
		}

		if err != nil {
			log.Printf("host %v request errored: %v", r.RemoteAddr, err)
			return
		}
	}
}
