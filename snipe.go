
/*
	https://github.com/gorilla/websocket/tree/master/examples/filewatch

	$ go get github.com/gorilla/websocket
	$ go run snipe.go
	# Open http://localhost:9090/ .
	# Type commands or text in the command prompt to see it in the browser
*/

package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"
	"bufio"
	"fmt"

	"github.com/gorilla/websocket"
)

// Time allowed to read the next pong message from the client.
const pongWait = 3600 * time.Second

var (
	addr      = flag.String("addr", ":9090", "http service address")
	homeTempl = template.Must(template.New("").Parse(homeHTML))
	upgrader  = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func reader(ws *websocket.Conn) {
	defer ws.Close()
	defer os.Exit(0)

	ws.SetReadLimit(512)
	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			break
		}
	}
}

func writer(ws *websocket.Conn) {
	for {
		fmt.Print("> ")
		bio := bufio.NewReader(os.Stdin)
		line, _, _ := bio.ReadLine()
		if err := ws.WriteMessage(websocket.TextMessage, []byte(line)); err != nil {
			return
		}
	}
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println(err)
		}
		return
	}

	go writer(ws)
	reader(ws)
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var v = struct {
		Host    string
		Data    string
	}{
		r.Host,
		defaultContent,
	}
	homeTempl.Execute(w, &v)
}

func main() {
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", serveWs)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal(err)
	}
}

const defaultContent = `<strong>commands</strong><br /><em>#cmd clear</em><br /><em>#cmd dump</em><br /><em>#cmd redrum</em><br /><br />`

const homeHTML = `<!DOCTYPE html>
<html lang="en">
  <head>
      <title>WebSocket Example</title>
  </head>
  <body id="body">
  	{{.Data}}
    <script src="http://ajax.googleapis.com/ajax/libs/jquery/2.1.1/jquery.min.js"></script>
    <script type="text/javascript">
    	var modes = { normal: 'normal', redrum: 'redrum' };
    	var mode = modes.normal;
      (function($) {
        var el = document.getElementById("body");
        var $el = $(el);
        var $html = $('html');
        var conn = new WebSocket("ws://{{.Host}}/ws");
        conn.onclose = function(evt) {
          el.textContent = 'Connection closed';
        }
        conn.onmessage = function(evt) {
      		if (!evt || !evt.data) return;

          if (evt.data.trim().toLowerCase() === '#cmd clear') {
          	el.innerHTML = '{{.Data}}';
          	$el.css({'background-color':'white', 'color':'black', 'font-size':16});
          	console.clear();
          	mode = modes.normal;
          } else if (evt.data.trim().toLowerCase() === '#cmd redrum' || mode === modes.redrum) {
          	if (mode !== modes.redrum) {
          		mode = modes.redrum;
          		el.innerHTML = 'REDRUM';
          		$el.css({'background-color':'red', 'color':'white', 'font-size':72});
          		console.clear();
          	} else {
							el.innerHTML += ' REDRUM';
          	}
          	
          	console.log('REDRUM');
          } else if (evt.data.trim().toLowerCase() === '#cmd dump') {
          	console.log($html.get(0).outerHTML);
          	mode = modes.normal;
          }
          else {
          	el.innerHTML += evt.data;
          	el.innerHTML += '<br />';
          }
        }
      })(jQuery);
    </script>
  </body>
</html>
`