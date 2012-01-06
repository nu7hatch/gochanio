package chanio

import (
	"encoding/gob"
	"io"
	"sync"
)

// packet is a wrapper for data passed over the io interfaces.
type packet struct {
	X interface{}
}

// Reader implements channel's binding for io.Reader.
type Reader struct {
	ch    chan<- interface{}
	r     io.Reader
	alive bool
	mtx   sync.Mutex
}

// NewReader returns a new Reader which reads data from specified
// io.Reader and proxifies it to Incoming channel.
//
// Example:
//
//     conn := net.Dial("tcp", "host.com:8080")
//     r := chanio.NewReader(conn, ch)
//     for x := range r.Incoming {
//         // do something with x
//     }
//
func NewReader(r io.Reader, ch chan<- interface{}) (chr *Reader) {
	chr = &Reader{ch: ch, r: r, alive: true}
	go chr.read()
	return
}

// Close terminates reading loop and closes the Incoming channel.
func (chr *Reader) Close() error {
	chr.mtx.Lock()
	defer chr.mtx.Unlock()
	chr.alive = false
	return nil
}

// read handles all the data read from the underlaying io.Reader
// and passes it to the Incoming channel.
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

// Writer implements channel's binding for io.Writer.
type Writer struct {
	ch   <-chan interface{}
	dead chan bool
	w    io.Writer
}

// NewWriter returns a new Writer which takes data from the Outgoing
// channel and proxifies it to specified io.Writer.
//
// Example:
//
//     conn := net.Dial("tcp", "host.com:8080")
//     r := chanio.NewReader(conn, ch)
//     for x := range r.Incoming {
//         // do something with x
//     }
//
func NewWriter(w io.Writer, ch <-chan interface{}) (chw *Writer) {
	chw = &Writer{ch: ch, dead: make(chan bool), w: w}
	go chw.write()
	return
}

// Close terminates reading loop and closes the Incoming channel.
func (chw *Writer) Close() error {
	chw.dead <- true
	return nil
}

// write handles all the data received from the Outgoing channel
// and writes it to the io.Writer. 
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
