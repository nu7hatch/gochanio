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
	ch := make(chan interface{})
	r := NewReader(buf, ch)
	x := <-ch
	val, ok := x.(int)
	if !ok || val != 1 {
		t.Errorf("Expected encoded value to be 1, given %s", val)
	}
	r.Close()
}

func TestWriter(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	ch := make(chan interface{})
	w := NewWriter(buf, ch)
	ch <- 1
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
	w.Close()
}

func TestReaderAndWriterOverTheNetwork(t *testing.T) {
	var wg sync.WaitGroup
	var r *Reader
	var w *Writer
	in, out := make(chan interface{}), make(chan interface{})
	host := "127.0.0.1:5678"

	wg.Add(1)
	go func() {
		addr, _ := net.ResolveTCPAddr("tcp", host)
		l, _ := net.ListenTCP("tcp", addr)
		for {
			var conn net.Conn
			conn, err := l.Accept()
			if err != nil {
				break
			}
			r = NewReader(conn, out)
			wg.Done()
		}
	}()

	conn, _ := net.Dial("tcp", host)
	w = NewWriter(conn, in)
	in <- 1
	x := <-out
	val, ok := x.(int)
	if !ok || val != 1 {
		t.Errorf("Expected to pass 1 over the network, given %s", val)
	}
	w.Close()
	r.Close()
}
