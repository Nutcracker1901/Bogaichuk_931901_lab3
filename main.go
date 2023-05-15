package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"flag"
)

type Node struct {
	id int
	left_chan <-chan Msg
	right_chan chan Msg
}

type Msg struct {
	Data string `json:"data"`
	Recv int `json:"recv"`
	Ttl int `json:"ttl"`
}

func (node *Node) Run() {
	msg := <-node.left_chan
	switch {
	case msg.Recv == node.id:
		fmt.Println("NodeID", node.id, "Recieve ( Recv:", msg.Recv, "Data:", msg.Data," Ttl:", msg.Ttl, ")")
	case msg.Ttl > 0:
		msg.Ttl -= 1
		node.right_chan <- msg
		//fmt.Println("NodeID", node.id, "Pass ( Recv:", msg.Recv, " Data:", msg.Data," Ttl:", msg.Ttl, ")")
	default:
		fmt.Println("NodeID", node.id, "Expired ( Recv:", msg.Recv, "Data:", msg.Data," Ttl:", msg.Ttl, ")")
	}
}

func initialize(n int) []*Node {
		ring := make([]*Node, 0, n)
		
		ring = append(ring, &Node{id:0, right_chan: make(chan Msg)})
		
		for i:=1; i < n; i++ {
			ring = append(ring, &Node{id:i, left_chan: ring[i-1].right_chan, right_chan: make(chan Msg)})
		}

		ring[0].left_chan = ring[n-1].right_chan

		return ring
		
}

func sendMsg(w http.ResponseWriter, r *http.Request) {
	msg := Msg{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&msg)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Println(msg.Data, " ", msg.Recv, "", msg.Ttl)

	for _, node := range ring {
		go node.Run()
 	}

 	ring[len(ring)-1].right_chan <- msg
}


var ring []*Node


func main() {
	var nodeCnt = flag.Int("n", 3, "Number of nodes")
	flag.Parse()

	ring = initialize(*nodeCnt)

	http.HandleFunc("/", sendMsg)

	err := http.ListenAndServe(":3333", nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}