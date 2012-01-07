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
//             break
//         }
//         r = chanio.NewReader(conn)
//         for x := range r.Incoming {
//             // do something with x
//         }
//     }
//
// Client:
//
//     conn, _ := net.Dial("tcp", host)
//     w = chanio.NewWriter(conn)
//     w.Outgoing <- "Hello World!"
//
package chanio
