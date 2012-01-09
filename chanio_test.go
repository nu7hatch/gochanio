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
	w <- 1
	x := <-r
	val, ok := x.(int)
	if !ok || val != 1 {
		t.Errorf("Expected to pass 1 over the network, given %s", val)
	}
	close(w)
	_, ok = <-r
	if ok {
		t.Errorf("Expected r to be closed")
	}
}

func TestExchangingBasicTypes(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	w := NewWriter(buf)
	w <- 1
	w <- "foo"
	w <- float32(1.23)
	r := NewReader(buf)
	i := <-r
	if val, ok := i.(int); !ok || val != 1 {
		t.Errorf("Expected to pass an int value, given %s", i)
	}
	s := <-r
	if val, ok := s.(string); !ok || val != "foo" {
		t.Errorf("Expected to pass a string value, given %s", s)
	}
	f := <-r
	if val, ok := f.(float32); !ok || val != 1.23 {
		t.Errorf("Expected to pass a float32 value, given %s", f)
	}
}

func TestExchangingSlices(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	w := NewWriter(buf)
	w <- []string{"foo", "bar"}
	r := NewReader(buf)
	x := <-r
	l, ok := x.([]string)
	if !ok {
		t.Errorf("Expected to pass a slice, given %s", x)
		return
	}
	if len(l) != 2 || l[0] != "foo" || l[1] != "bar" {
		t.Errorf("Expected to pass a slice with correct values, given %s", l)
	}
}

func TestExchangingMaps(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	w := NewWriter(buf)
	w <- map[string]int{"foo": 1}
	r := NewReader(buf)
	x := <-r
	m, ok := x.(map[string]int)
	if !ok {
		t.Errorf("Expected to pass a map, given %s", x)
		return
	}
	if foo, ok := m["foo"]; !ok || foo != 1 {
		t.Errorf("Expected to pass a map with correct values, given %s", m)
	}
}

func TestExchangingStructs(t *testing.T) {
	gob.Register(new(struct{A int}))
	buf := bytes.NewBuffer([]byte{})
	w := NewWriter(buf)
	w <- &struct{A int}{1}
	r := NewReader(buf)
	x := <-r
	s, ok := x.(*struct{A int})
	if !ok {
		t.Errorf("Expected to pass a struct, given %s", x)
		return
	}
	if s.A != 1 {
		t.Errorf("Expected to pass a struct with correct values, given %s", s)
	}
}