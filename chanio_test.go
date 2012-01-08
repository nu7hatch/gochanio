package chanio

import (
	"bytes"
	"encoding/gob"
	"net"
	"sync"
	"testing"
)

func TestReader(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	enc.Encode(packet{1})
	r := NewReader(buf)
	x := <-r
	val, ok := x.(int)
	if !ok || val != 1 {
		t.Errorf("Expected encoded value to be 1, given %s", val)
	}
}

func TestWriter(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	w := NewWriter(buf)
	w <- 1
	dec := gob.NewDecoder(buf)
	var p packet
	err := dec.Decode(&p)
	if err != nil {
		t.Errorf("Expected to decode a packet, error: %v", err)
	}
	val, ok := p.X.(int)
	if !ok || val != 1 {
		t.Errorf("Expected decoded value to be 1, given %s", val)
	}
	close(w)
}

func TestReaderAndWriterOverTheNetwork(t *testing.T) {
	var wg sync.WaitGroup
	var r <-chan interface{}
	var w chan<- interface{}
	host := "127.0.0.1:5678"

	addr, _ := net.ResolveTCPAddr("tcp", host)
	l, _ := net.ListenTCP("tcp", addr)
	wg.Add(1)
	go func() {
		conn, err := l.Accept()
		if err != nil {
			t.Errorf("Expected to accept connection, error: %v", err)
		}
		r = NewReader(conn)
		wg.Done()
	}()

	conn, _ := net.Dial("tcp", host)
	w = NewWriter(conn)
	wg.Wait()
	println("ok")
	w <- 1
	x := <-r
	val, ok := x.(int)
	if !ok || val != 1 {
		t.Errorf("Expected to pass 1 over the network, given %s", val)
	}
	close(w)
	conn.Close()
}
