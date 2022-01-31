package server

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	// "golang.org/x/time/rate"
	"nhooyr.io/websocket"
)

// Implements a server with basic authentication and request rate
// Limiting. Only supports websockets connection (wss:// only)
type BasicServer struct {
	// Task's entry point. Since the request format is literally a command
	// specified on the command-line, then this should suffice for now
	TaskEntryPoint func(calledFromRepl bool)
}

func (b *BasicServer) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", b.serveWSS)

	srv := &http.Server{
		Addr:         "localhost:9090",
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	log.Printf("starting websockets server on %s", srv.Addr)
	// Use FiloSottile's [mkcert](https://github.com/FiloSottile/mkcert) utility
	err := srv.ListenAndServeTLS("./localhost.pem", "./localhost-key.pem")
	// err := srv.ListenAndServe()
	if err != nil {
		log.Fatal("server error ", err)
		return err
	}
	return nil

}

func (b *BasicServer) startExecution(ctx context.Context, c *websocket.Conn) error {

	// msgType, r, err := c.Reader(ctx)
	// if err != nil {
	// 	log.Println("reader err")
	// 	return err
	// }
	//
	// w, err := c.Writer(ctx, msgType)
	// if err != nil {
	// 	log.Println("writer err")
	// 	return err
	// }
	//
	var cmd []byte
	_, cmd, err := c.Read(ctx)
	if err != nil {
		log.Fatal("read error")
		if err == io.EOF {
			return nil
		}
		return err
	}
	// log.Printf("read command %s of message type %s ", cmd, msgType)
        // TODO: input is treated like a REPL command. this should change
        // to a JSON API that slightly exposes Task
	args := strings.Split(string(cmd), " ") // naiive. remove me
	os.Args = append([]string{"task"}, args...)
	b.TaskEntryPoint(true)

	return nil

}

// Listens for wss:// connections
// Example modified from the official documentation for nhooyr.io/websocket
// https://github.com/nhooyr/websocket/blob/v1.8.7/examples/echo/server.go
func (b *BasicServer) serveWSS(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{})
	// 	Subprotocols: []string{"echo"},
	// })
	if err != nil {
		log.Fatalf("%v", err)
		return
	}
	defer c.Close(websocket.StatusInternalError, "internal error")

	// if c.Subprotocol() != "echo" {
	// 	c.Close(websocket.StatusPolicyViolation, "client must speak the echo subprotocol")
	// 	return
	// }

	// l := rate.NewLimiter(rate.Every(time.Millisecond*100), 10)
	for {
		err = b.startExecution(r.Context(), c)
		if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
			return
		}
		if err != nil {
			log.Fatalf("host %v request errored: %v", r.RemoteAddr, err)
			return
		}
	}
}
