package chanio

import (
	"encoding/gob"
	"io"
	"sync"
)

type packet struct {
	X interface{}
}

type Reader struct {
	ch    chan<- interface{}
	r     io.Reader
	alive bool
	mtx   sync.Mutex
}

func NewReader(r io.Reader, ch chan<- interface{}) (chr *Reader) {
	chr = &Reader{ch: ch, r: r, alive: true}
	go chr.read()
	return
}

func (chr *Reader) Close() error {
	chr.mtx.Lock()
	defer chr.mtx.Unlock()
	chr.alive = false
	return nil
}

func (chr *Reader) read() {
	dec := gob.NewDecoder(chr.r)
	for {
		if !chr.alive {
			break
		}
		var e packet
		err := dec.Decode(&e)
		if err != nil {
			if err == io.EOF {
				break
			}
			continue
		}
		chr.ch <- e.X
	}
}

type Writer struct {
	ch   <-chan interface{}
	dead chan bool
	w    io.Writer
}

func NewWriter(w io.Writer, ch <-chan interface{}) (chw *Writer) {
	chw = &Writer{ch: ch, dead: make(chan bool), w: w}
	go chw.write()
	return
}

func (chw *Writer) Close() error {
	chw.dead <- true
	return nil
}

func (chw *Writer) write() {
	enc := gob.NewEncoder(chw.w)
	for {
		select {
		case <-chw.dead:
			return
		case x := <-chw.ch:
			err := enc.Encode(&packet{x})
			if err != nil {
				if err == io.EOF {
					return
				}
				continue
			}
		}
	}
}
