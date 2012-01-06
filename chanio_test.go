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
	x := <-r.Incoming
	val, ok := x.(int)
	if !ok || val != 1 {
		t.Errorf("Expected encoded value to be 1, given %s", val)
	}
	r.Close()
}

func TestWriter(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	w := NewWriter(buf)
	w.Outgoing <- 1
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
	host := "127.0.0.1:5678"

	wg.Add(1)
	go func() {
		addr, _ := net.ResolveTCPAddr("tcp", host)
		l, _ := net.ListenTCP("tcp", addr)
		for {
			conn, err := l.Accept()
			if err != nil {
				break
			}
			r = NewReader(conn)
			wg.Done()
		}
	}()

	conn, _ := net.Dial("tcp", host)
	w = NewWriter(conn)
	wg.Wait()
	w.Outgoing <- 1
	x := <-r.Incoming
	val, ok := x.(int)
	if !ok || val != 1 {
		t.Errorf("Expected to pass 1 over the network, given %s", val)
	}
	w.Close()
	r.Close()
}
