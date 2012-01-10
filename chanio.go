// Package chanio provides Reader and Writer bindings with
// go channels for io.Reader and io.Writer interfaces.
//
// Here's an example implementation of channels communication
// over the network:
//
// Server:
//
//     addr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:8080")
//     l, _ := net.ListenTCP("tcp", addr)
//     for {
//         conn, err := l.Accept()
//         if err != nil {
//             continue
//         }
//         r := chanio.NewReader(conn)
//         for x := range r {
//             // do something with x
//         }
//     }
//
// Client:
//
//     conn, _ := net.Dial("tcp", host)
//     w := chanio.NewWriter(conn)
//     w <- "Hello World!"
//
// The chanio package is using encoding/gob to encode and decode
// data exchanged over the underlying buffers. If you want to send
// a struct, map or value of a custom type, then you need to register
// it in gob first. Example:
//
//     package main
//
//     import (
//         "encoding/gob"
//         "chanio"
//         "net"
//     )
//
//     type Cat struct {
//         Name   string
//         IsCute bool
//     }
// 
//     func main() {
//         gob.Register(&Cat{})
//         conn, _ := net.Dial("tcp", "host.com:8080")
//         w := chanio.NewWriter(conn)
//         w <- &Cat{Name: "Tom", IsCute: false}
//         conn.Close()
//     }
//
package chanio

import (
	"encoding/gob"
	"io"
)

// packet is a wrapper for data passed over the io interfaces.
type packet struct {
	X interface{}
}

// NewReader returns a new read-only channel which provides data
// read from specified io.Reader.
//
// Example:
//
//     conn := net.Dial("tcp", "host.com:8080")
//     r := chanio.NewReader(conn)
//     for x := range r {
//         // do something with x
//     }
//
func NewReader(r io.Reader) <-chan interface{} {
	ch := make(chan interface{})
	go read(r, ch)
	return ch
}

// read handles all the data read from the underlaying io.Reader
// and passes it to the specified channel.
func read(r io.Reader, ch chan interface{}) {
	defer close(ch)
	dec := gob.NewDecoder(r)
	for {
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

// NewWriter returns a new write-only channel which passes data
// to specified io.Writer.
//
// Example:
//
//     conn := net.Dial("tcp", "host.com:8080")
//     w := chanio.NewWriter(conn)
//     w <- "foo"
//
func NewWriter(w io.Writer) chan<- interface{} {
	ch := make(chan interface{})
	go write(w, ch)
	return ch
}

// write handles all the data received from specified channel
// and writes it to the io.Writer. 
func write(w io.Writer, ch chan interface{}) {
	enc := gob.NewEncoder(w)
	for x := range ch {
		err := enc.Encode(&packet{x})
		if err != nil {
			if err == io.EOF {
				break
			}
			continue
		}
	}
	if wc, ok := w.(io.WriteCloser); ok {
		wc.Close()
	}
}
