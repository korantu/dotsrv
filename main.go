/**
Plan:
1. Web server, To run 3d.js, sort-of done.
2. Parse csv.
3. Nodes and links, look like hash table.
  - Name all the IPs'
  - Then hashtable.
*/
package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// serve does file system serving, for testing web part
func serve() {
	log.Print("Serving at port 8080")

	if err := http.ListenAndServe(":8080", http.FileServer(http.Dir("."))); err != nil {
		log.Print("Unable To serve: ", err.Error())
	} else {
		log.Print("All is ok.")
	}
}

// Keeping track of events
type Event struct {
	From, To string
	Type     string
}

type Events []Event

// Test data
var data string = `
Unauthorized  Users on Cardholder Objects - Alert,Guardium @ 10.0.3.60,1,2010-10-06 11:26:08,Unauthorized Access Attempt,10.0.240.52,22109,10.0.10.41,120,mabel_santoro,8
Unauthorized  Users on Cardholder Objects - Alert,Guardium @ 10.0.3.60,1,2010-10-06 11:26:08,Unauthorized Access Attempt,10.0.240.52,22109,10.0.10.41,120,mabel_santoro,8
Unauthorized  Users on Cardholder Objects - Alert,Guardium @ 10.0.3.60,1,2010-10-06 11:26:08,Unauthorized Access Attempt,10.0.240.52,22109,10.0.10.41,120,mabel_santoro,8
Unauthorized  Users on Cardholder Objects - Alert,Guardium @ 10.0.3.60,1,2010-10-06 11:26:08,Unauthorized Access Attempt,10.0.240.52,22109,10.0.10.41,120,mabel_santoro,8
Unauthorized  Users on Cardholder Objects - Alert,Guardium @ 10.0.3.60,1,2010-10-06 11:26:08,Unauthorized Access Attempt,10.0.240.52,22109,10.0.10.41,120,mabel_santoro,8
Unauthorized  Users on Cardholder Objects - Alert,Guardium @ 10.0.3.60,1,2010-10-06 11:26:08,Unauthorized Access Attempt,10.0.240.52,22109,10.0.10.41,120,mabel_santoro,8
`

func testData() io.Reader {
	return strings.NewReader(data)
}

type Edge struct {
	From, To int
}

type EdgeId int

func getEdgeId(an Edge) EdgeId {
	return EdgeId(an.From + (an.To << 14))
}

type Graph struct {
	NodeIndex map[string]int // Get from full name to index of the host.
	Nodes     []string       // Get from index to the name of the host.

	KnownEdges map[EdgeId]struct{} // Lookup for edges already in the system.
	Edges      []Edge
}

func NewGraph() Graph {
	return Graph{map[string]int{}, []string{}, map[EdgeId]struct{}{}, []Edge{}}
}

// addLink
func (a *Graph) addLink(from, to string) {

	if from == to {
		return
	}

	isInternal := func(addr string) bool {
		log.Print(addr)
		return strings.HasPrefix(addr, "10.0") || strings.Contains(addr, "_")
	}

	if isInternal(from) && isInternal(to) {
		// Internal, interesting.
	} else {
		return
	}

	getIndex := func(some string) int {
		if n, ok := a.NodeIndex[some]; ok {
			return n
		}
		// Not found; Add first.
		idx := len(a.Nodes)
		a.Nodes = append(a.Nodes, some)
		a.NodeIndex[some] = idx
		log.Printf("%d: %s", idx, some)
		return idx
	}

	ne := Edge{getIndex(from), getIndex(to)}
	ne_id := getEdgeId(ne)

	if _, ok := a.KnownEdges[ne_id]; !ok {
		a.KnownEdges[ne_id] = struct{}{}
		a.Edges = append(a.Edges, ne)
	}
}

// Process a set of events and derive graph data
func Process(some Events) Graph {
	g := NewGraph()

	for _, ev := range some {
		if strings.Contains(ev.Type, "Login") {
			g.addLink(ev.From, ev.To)
		}
	}

	return g
}

func DumpTgf(f io.Writer, a Graph) {

	for i, _ := range a.Nodes {
		fmt.Fprintf(f, "%d %d\n", i, i)
	}

	fmt.Fprintln(f, "#")

	for _, e := range a.Edges {
		fmt.Fprintf(f, "%d %d\n", e.From, e.To)
	}

}

func DumpBasic(a Graph) {
	entries := map[string]struct{}{}
	for _, e := range a.Edges {
		entry := fmt.Sprint(e.From, "->", e.To)
		if _, ok := entries[entry]; !ok {
			fmt.Println(entry)
		}
		entries[entry] = struct{}{}
	}
}

// readEvents gets raw data input and derives events
func readEvents(an io.Reader) Events {
	evts := Events{}

	scn := bufio.NewScanner(an)
	scn.Split(bufio.ScanLines)
	for scn.Scan() {
		line := string(scn.Bytes())
		if strings.Contains(line, ",") {
			items := strings.Split(line, ",")
			max := len(items)
			if max > 7 {
				evts = append(evts, Event{items[max-6], items[max-4], items[0]})
			}
		}
	}

	return evts

}

func main() {
	//serve()
	// f, err := os.Open(`C:\log\atoms\gosrv\data\2010-10-06-data_export.csv`)
	f, err := os.Open(`C:\log\atoms\gosrv\data\2010-10-07-data_export.csv`)
	if err != nil {
		log.Fatal("Problem opening data file: ", err.Error())
	}
	defer f.Close()

	//evts := readEvents(testData())
	evts := readEvents(f)

	gr := Process(evts)
	log.Printf("Logins: %d\n", len(gr.Nodes))

	out := "Logins.tgf"
	log.Printf("Dumping to %s:", out)
	fout, err := os.Create(out)
	if err != nil {
		log.Fatal("Problem opening data file: ", err.Error())
	}
	
	defer fout.Close()
	log.Println("Ok.")

	DumpTgf(fout, gr)
	//	DumpBasic(gr)
}














