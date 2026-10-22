package other

import (
	"net"
	"testing"
	"time"
)

func TestSMTPClientTimesOutWhenServerStalls(t *testing.T) {
	for _, encrypted := range []bool{false, true} {
		t.Run(map[bool]string{false: "greeting", true: "tls-handshake"}[encrypted], func(t *testing.T) {
			listener, err := net.Listen("tcp", "localhost:0")
			if err != nil {
				t.Fatal(err)
			}
			defer listener.Close()
			stop := make(chan struct{})
			defer close(stop)
			go func() {
				conn, acceptErr := listener.Accept()
				if acceptErr != nil {
					return
				}
				defer conn.Close()
				<-stop
			}()
			started := time.Now()
			client, err := newSMTPClient(listener.Addr().String(), encrypted, 30*time.Millisecond)
			if client != nil {
				client.Close()
			}
			if err == nil || time.Since(started) > time.Second {
				t.Fatalf("stalled SMTP did not time out: %v", err)
			}
		})
	}
}
