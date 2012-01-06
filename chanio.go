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
	Incoming <-chan interface{}
	r        io.Reader
	closed   bool
	mtx      sync.Mutex
}

// NewReader returns a new Reader which reads data from specified
// io.Reader and proxifies it to Incoming channel.
//
// Example:
//
//     conn := net.Dial("tcp", "host.com:8080")
//     r := chanio.NewReader(conn)
//     for x := range r.Incoming {
//         // do something with x
//     }
//
func NewReader(r io.Reader) (chr *Reader) {
	chr = &Reader{r: r, closed: false}
	ch := make(chan interface{})
	chr.Incoming = ch
	go chr.read(ch)
	return
}

// Close terminates reading loop and closes the Incoming channel.
func (chr *Reader) Close() error {
	chr.mtx.Lock()
	defer chr.mtx.Unlock()
	chr.closed = true
	return nil
}

// isClosed returns false if reader has been closed.
func (chr *Reader) isClosed() bool {
	chr.mtx.Lock()
	defer chr.mtx.Unlock()
	return chr.closed
}

// read handles all the data read from the underlaying io.Reader
// and passes it to the Incoming channel.
func (chr *Reader) read(ch chan interface{}) {
	dec := gob.NewDecoder(chr.r)
	for {
		if chr.isClosed() {
			close(ch)
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
		ch <- e.X
	}
}

// Writer implements channel's binding for io.Writer.
type Writer struct {
	Outgoing chan<- interface{}
	ch       chan interface{}
	w        io.Writer
}

// NewWriter returns a new Writer which takes data from the Outgoing
// channel and proxifies it to specified io.Writer.
//
// Example:
//
//     conn := net.Dial("tcp", "host.com:8080")
//     w := chanio.NewWriter(conn)
//     w.Outgoing <- "foo"
//
func NewWriter(w io.Writer) (chw *Writer) {
	chw = &Writer{w: w}
	chw.ch = make(chan interface{})
	chw.Outgoing = chw.ch
	go chw.write()
	return
}

// Close terminates reading loop and closes the Incoming channel.
func (chw *Writer) Close() error {
	close(chw.ch)
	return nil
}

// write handles all the data received from the Outgoing channel
// and writes it to the io.Writer. 
func (chw *Writer) write() {
	enc := gob.NewEncoder(chw.w)
	for x := range chw.ch {
		err := enc.Encode(&packet{x})
		if err != nil {
			if err == io.EOF {
				break
			}
			continue
		}
	}
}
