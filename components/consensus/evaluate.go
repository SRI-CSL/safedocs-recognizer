package main

import (
	"container/list"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	iradix "github.com/hashicorp/go-immutable-radix"

	"github.com/awalterschulze/gographviz"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	args := os.Args[1:]
	cfgDiffMode := args[0] == "cfg-diff"
	dumpMode := args[0] == "dump"
	baselineParser := ""
	testParser := ""
	docFilter := ""
	if cfgDiffMode && baselineParser == "" && testParser == "" {
		log.Fatal("baseline parser and test parser are needed as args")
	}
	if cfgDiffMode && len(args) == 4 {
		baselineParser = args[1]
		testParser = args[2]
		docFilter = args[3]
	} else if cfgDiffMode && len(args) == 3 {
		baselineParser = args[1]
		testParser = args[2]
	} else if (dumpMode) && len(args) == 2 {
		docFilter = args[1]
	}

	pgConn := "user=postgres password=postgres host='127.0.0.1' sslmode=disable"
	if len(os.Getenv("MR_POSTGRES_CONN")) > 0 {
		pgConn = os.Getenv("MR_POSTGRES_CONN")
	}
	if cfgDiffMode {
		parserCfgDiff(pgConn, baselineParser, testParser, docFilter)
	} else if dumpMode {
		dump(pgConn, docFilter)
	}
}

func dump(connStr string, query string) {
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	rows, err := db.Queryx(dumpQuery, "%"+query+"%")
	if err != nil {
		log.Fatal(err)
	}

	// dir, err := ioutil.TempDir("", "recognizer_"+query+"_")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	basedir := "./dump/"
	dir := basedir + query
	err = os.MkdirAll(dir, 0777)
	if err != nil {
		log.Fatal(err)
	}

	result := Result{}
	for rows.Next() {
		err := rows.StructScan(&result)
		if err != nil {
			log.Fatal(err)
		}
		tmpfncallgrind := filepath.Join(dir, result.Filename+"."+result.Parser+".callgrind")
		if err := ioutil.WriteFile(tmpfncallgrind, []byte(result.Callgrind), 0666); err != nil {
			log.Fatal(err)
		}
		tmpfncfg := filepath.Join(dir, result.Filename+"."+result.Parser+".cfg")
		if err := ioutil.WriteFile(tmpfncfg, []byte(result.Cfg), 0666); err != nil {
			log.Fatal(err)
		}
		tmpfncfgimage := filepath.Join(dir, result.Filename+"."+result.Parser+".cfg.png")
		if err := ioutil.WriteFile(tmpfncfgimage, result.CfgImage, 0666); err != nil {
			log.Fatal(err)
		}
		tmpfnstdout := filepath.Join(dir, result.Filename+"."+result.Parser+".stdout.txt")
		if err := ioutil.WriteFile(tmpfnstdout, []byte(result.Stdout), 0666); err != nil {
			log.Fatal(err)
		}
		tmpfnstderr := filepath.Join(dir, result.Filename+"."+result.Parser+".stderr.txt")
		if err := ioutil.WriteFile(tmpfnstderr, []byte(result.Stderr), 0666); err != nil {
			log.Fatal(err)
		}
		tmpfnstatus := filepath.Join(dir, result.Filename+"."+result.Parser+"."+result.Status)
		if err := ioutil.WriteFile(tmpfnstatus, []byte(result.Status), 0666); err != nil {
			log.Fatal(err)
		}
		// "dot -o callgrind.png -T png callgrind.dot"
		// need to call dot now that it isn't stored in the DB
		tmpfnpng := filepath.Join(dir, result.Filename+"."+result.Parser+".png")
		dot := exec.Command("dot", "-o", tmpfnpng, "-T", "png", tmpfncfg)
		err = dot.Run()
		if err != nil {
			log.Fatal(err)
		}
	}
	log.Print("dump saved to " + dir)

	// cmd := exec.Command("cp", "viewer/grid.css", basedir)
	// err = cmd.Run()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// cmd = exec.Command("cp", "viewer/index.html", basedir)
	// err = cmd.Run()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// cmd = exec.Command("cp", "viewer/script.js", basedir)
	// err = cmd.Run()
	// if err != nil {
	// 	log.Fatal(err)
	// }
}

type stats struct {
	Name       string
	Percentage []float64
	CallCount  []float64
	Min        int
	Quartile1  int
	Median     int
	Quartile3  int
	Max        int
}

func generateBaselineCfg(connStr string, parser string) *gographviz.Graph {
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	combinedGraphAst, _ := gographviz.ParseString(`digraph G {}`)
	combinedGraph := gographviz.NewGraph()
	if err := gographviz.Analyse(combinedGraphAst, combinedGraph); err != nil {
		panic(err)
	}

	rows, err := db.Queryx(parserQuery, parser, true)
	if err != nil {
		log.Fatal(err)
	}

	result := Result{}
	for rows.Next() {
		err := rows.StructScan(&result)
		if err != nil {
			log.Fatal(err)
		}
		if len(result.Cfg) == 0 {
			continue
		}
		graphAst, _ := gographviz.ParseString(result.Cfg)
		graph := gographviz.NewGraph()
		if err := gographviz.Analyse(graphAst, graph); err != nil {
			log.Fatal(err)
		}

		for _, n := range graph.Nodes.Nodes {
			n.Attrs["label"] = n.Name
			combinedGraph.Nodes.Add(n)
		}
		for _, e := range graph.Edges.Edges {
			if combinedGraph.Edges.SrcToDsts[e.Src][e.Dst] == nil {
				combinedGraph.AddEdge(e.Src, e.Dst, true, nil)
			}
		}
	}
	defer rows.Close()

	return combinedGraph
}

func getPaths(nSrc string, nDest string, g *gographviz.Graph) *list.List {
	knownPaths := list.New()
	n1 := g.Nodes.Lookup[nSrc]
	n2 := g.Nodes.Lookup[nDest]
	if n1 == nil || n2 == nil {
		return knownPaths
	}
	visited := make(map[string]*gographviz.Node)
	path := list.New()
	path.PushBack(n1)
	q := list.New()
	q.PushBack(path)
	visited[n1.Name] = n1

	for q.Len() > 0 {
		front := q.Front()
		path = front.Value.(*list.List)
		q.Remove(front)
		lastNode := path.Back().Value.(*gographviz.Node)

		if lastNode.Name == n2.Name {
			knownPaths.PushBack(path)
		}

		friends := g.Edges.SrcToDsts[lastNode.Name]

		for dst := range friends {
			if _, ok := visited[dst]; !ok {
				dstNode := g.Nodes.Lookup[dst]
				newpath := list.New()
				newpath.PushBackList(path)
				newpath.PushBack(dstNode)
				visited[dst] = dstNode
				q.PushBack(newpath)
			}
		}
	}

	return knownPaths
}

func bfs(n *gographviz.Node, g *gographviz.Graph) []*gographviz.Node {
	visited := make(map[string]*gographviz.Node)
	queue := list.New()
	queue.PushBack(n)
	visited[n.Name] = n

	for queue.Len() > 0 {
		qnode := queue.Front()

		friends := g.Edges.SrcToDsts[qnode.Value.(*gographviz.Node).Name]

		for dst := range friends {
			if _, ok := visited[dst]; !ok {
				dstNode := g.Nodes.Lookup[dst]
				visited[dst] = dstNode
				queue.PushBack(dstNode)
			}
		}
		queue.Remove(qnode)
	}

	nodes := make([]*gographviz.Node, 0)
	for _, node := range visited {
		nodes = append(nodes, node)
	}

	return nodes
}

func makePathString(p *list.List) string {
	var pathString string
	for n := p.Front(); n != nil; n = n.Next() {
		node := n.Value.(*gographviz.Node)
		pathString += node.Name
	}
	return pathString
}

func isSubtree(allPaths *list.List, thisPath string) bool {
	subtree := false
	for e := allPaths.Front(); e != nil; e = e.Next() {
		if subtree {
			break
		}
		p := e.Value.(*list.List)
		pathString := makePathString(p)
		if strings.Contains(pathString, thisPath) {
			subtree = true
		}
		if strings.Contains(thisPath, pathString) {
			subtree = true
		}
	}
	return subtree
}

func parserCfgDiff(connStr string, baselineParser string, testParser string, docFilter string) {
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	// var parsers []string
	// rows, err := db.Queryx(listParsersQuery)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// for rows.Next() {
	// 	var p string
	// 	err := rows.Scan(&p)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	parsers = append(parsers, p)
	// }
	// rows.Close()

	// parsers = append(parsers, "poppler")
	// parsers = append(parsers, "qpdf")
	// parsers = append(parsers, baselineParser)

	filesWithNewPaths := 0
	filesWithNewDsts := 0

	var combinedGraph *gographviz.Graph
	combinedGraphFile := baselineParser + "_combined_cfg.dot"
	combinedGraphBytes, err := ioutil.ReadFile(combinedGraphFile)
	if err != nil {
		combinedGraph = generateBaselineCfg(connStr, baselineParser)
		err = ioutil.WriteFile(combinedGraphFile, []byte(combinedGraph.String()), 0666)
		if err != nil {
			log.Fatal(err)
		}
		// log.Fatal("wrote file, stopping, good")
	} else {
		graphAst, _ := gographviz.ParseString(string(combinedGraphBytes))
		combinedGraph = gographviz.NewGraph()
		err := gographviz.Analyse(graphAst, combinedGraph)
		if err != nil {
			log.Fatal(err)
		}
	}

	var rows *sqlx.Rows
	if docFilter == "" {
		rows, err = db.Queryx(parserQuery, testParser, false)
	} else {
		rows, err = db.Queryx(parserQueryDocFilter, testParser, false, "%"+docFilter+"%")
	}
	if err != nil {
		log.Fatal(err)
	}

	result := Result{}
	for rows.Next() {
		err := rows.StructScan(&result)
		if err != nil {
			log.Fatal(err)
		}

		graphAst, _ := gographviz.ParseString(result.Cfg)
		graph := gographviz.NewGraph()
		if err := gographviz.Analyse(graphAst, graph); err != nil {
			log.Fatal(err)
		}

		matchedSrcDstNewPaths := list.New()
		allSrcDstBaselinePaths := list.New()
		missingDstPaths := list.New()
		matchedPathCount := 0
		missingDstPathCount := 0

		rTree := iradix.New()

		// for every node in the parser run, maybe only root?
		for _, n := range graph.Nodes.Nodes {

			// if !strings.Contains(n.Name, "(below main)") {
			// 	continue
			// }
			// if !strings.Contains(n.Name, "1090") {
			// 	continue
			// }

			// get the same node from the baseline
			nbaseline := combinedGraph.Nodes.Lookup[n.Name]
			if nbaseline != nil {
				root := graph.Nodes.Lookup[n.Name]
				// get a list of all reachable nodes from this node
				reachable := bfs(root, graph)
				for _, r := range reachable {
					if strings.Contains(r.Name, "4e3c940") {
						continue
					}
					reachablePaths := list.New()
					baselinePaths := getPaths(root.Name, r.Name, combinedGraph)
					paths := getPaths(root.Name, r.Name, graph)
					if baselinePaths.Len() == 0 {
						// dst does not exist, all paths are new
						for e := paths.Front(); e != nil; e = e.Next() {
							thisPath := makePathString(e.Value.(*list.List))
							subtree := isSubtree(missingDstPaths, thisPath)
							if !subtree {
								missingDstPaths.PushBack(e.Value.(*list.List))
								missingDstPathCount++
							}
						}
					}
					// add path
					for e := paths.Front(); e != nil; e = e.Next() {
						thisPath := makePathString(e.Value.(*list.List))
						subtree := isSubtree(reachablePaths, thisPath)
						accountedFor := false
						for e1 := missingDstPaths.Front(); e1 != nil; e1 = e.Next() {
							if e1 == e {
								accountedFor = true
								break
							}
						}

						if !subtree && !accountedFor {
							reachablePaths.PushBack(e.Value.(*list.List))
						}
					}

					// remove path if it exists in the baseline (not a miss)
					for e := reachablePaths.Front(); e != nil; e = e.Next() {
						// fmt.Println(e.Value)
						thisPath := makePathString(e.Value.(*list.List))

						inBaseline := false
						for b := baselinePaths.Front(); b != nil; b = b.Next() {
							baselinePathString := makePathString(b.Value.(*list.List))
							if strings.Contains(baselinePathString, thisPath) || strings.Contains(thisPath, baselinePathString) {
								inBaseline = true
							}
						}

						if inBaseline {
							matchedPathCount++
						} else {
							// add baseline paths
							for e1 := matchedSrcDstNewPaths.Front(); e1 != nil; e1 = e.Next() {
								existingPath := makePathString(e1.Value.(*list.List))
								if strings.Contains(existingPath, thisPath) {
									// subpath, ignore
								} else if strings.Contains(thisPath, existingPath) {
									matchedSrcDstNewPaths.Remove(e1)
									matchedSrcDstNewPaths.PushBack(e.Value.(*list.List))
								} else {
									matchedSrcDstNewPaths.PushBack(e.Value.(*list.List))
								}
							}

							// rTree, _, _ = rTree.Insert([]byte(thisPath), 1)
						}
					}

					for e := baselinePaths.Front(); e != nil; e = e.Next() {
						allSrcDstBaselinePaths.PushBack(e.Value.(*list.List))
					}
				}
			}
		}

		var out []string
		fn := func(k []byte, v interface{}) bool {
			out = append(out, string(k))
			return false
		}
		rTree.Root().Walk(fn)
		// rTree.Root().WalkPrefix([]byte("main"), fn)
		// rTree.Root().LongestPrefix()
		// if len(out) > 0 {
		// 	fmt.Println(len(out))
		// }
		// for i := 0; i < len(out); i++ {
		// 	fmt.Println(out[i])
		// }
		// if len(out) > 1 {
		// 	fmt.Println()
		// }
		// if len(out) > 0 {
		// 	fmt.Println("=====")
		// }

		// if len(out) > 0 {
		// 	// fmt.Println(rTree.Root().Se([]byte(n.Name)))
		// 	fmt.Println(rTree.Root().LongestPrefix([]byte(n.Name)))
		// 	fmt.Println()
		// }

		if missingDstPathCount != missingDstPaths.Len() {
			log.Fatal("calculation error")
		}

		fmt.Println()
		fmt.Println(result.Filename)
		fmt.Println("matched paths: " + strconv.Itoa(matchedPathCount))
		fmt.Println("missed paths: " + strconv.Itoa(matchedSrcDstNewPaths.Len()))
		fmt.Println("missing dst (subset of missed paths) path count: " + strconv.Itoa(missingDstPathCount))
		fmt.Print()

		if missingDstPaths.Len() > 0 {
			filesWithNewDsts++
		}
		for a := missingDstPaths.Front(); a != nil; a = a.Next() {
			fmt.Println()
			fmt.Println("==========")
			fmt.Println(result.Filename)
			fmt.Println("======dst not in baseline, new path by default======")
			p := a.Value.(*list.List)
			for e := p.Front(); e != nil; e = e.Next() {
				fmt.Println(e.Value.(*gographviz.Node).Name)
			}
			fmt.Print()
		}

		if matchedSrcDstNewPaths.Len() > 0 {
			filesWithNewPaths++
		}
		for a := matchedSrcDstNewPaths.Front(); a != nil; a = a.Next() {
			fmt.Println()
			fmt.Println("==========")
			fmt.Println(result.Filename)
			fmt.Println("======src,dst reachable in baseline, new path======")
			p := a.Value.(*list.List)
			for e := p.Front(); e != nil; e = e.Next() {
				fmt.Println(e.Value.(*gographviz.Node).Name)
			}
		}

		fmt.Println("==========")
		fmt.Println()
	}
	rows.Close()

	fmt.Println("Files with new dsts: " + strconv.Itoa(filesWithNewDsts))
	fmt.Println("files with new path (src to dst): " + strconv.Itoa(filesWithNewPaths))
}

var reportQuery = `
SELECT substring(c.doc from '(?:.+/)(.+)') AS filename, 
	   c.doc, parser, status, stderr, digest  
FROM consensus c
WHERE baseline = $1
GROUP BY c.doc, parser, status, digest, stderr
ORDER BY c.doc, status
`

var reportQueryDocFilter = `
SELECT substring(c.doc from '(?:.+/)(.+)') AS filename, 
	   c.doc, parser, status, stderr, digest  
FROM consensus c
WHERE baseline = $1 AND doc LIKE $2
GROUP BY c.doc, parser, status, digest, stderr
ORDER BY c.doc, status
`

var dumpQuery = "SELECT substring(doc from '(?:.+/)(.+)') AS filename, parser, status, stdout, stderr, digest, callgrind, cfg, cfg_image " +
	"FROM consensus " +
	"WHERE doc LIKE $1 " +
	"GROUP BY doc, parser, status, digest, stderr, stdout, callgrind, cfg, cfg_image " +
	"ORDER BY doc, status"

var parserQuery = "SELECT substring(doc from '(?:.+/)(.+)') AS filename, parser, status, stdout, stderr, digest, callgrind, cfg, cfg_image " +
	"FROM consensus " +
	"WHERE parser = $1 AND baseline = $2" +
	"GROUP BY doc, parser, status, digest, stderr, stdout, callgrind, cfg, cfg_image " +
	"ORDER BY doc, status"

var parserQueryDocFilter = "SELECT substring(doc from '(?:.+/)(.+)') AS filename, parser, status, stdout, stderr, digest, callgrind, cfg, cfg_image " +
	"FROM consensus " +
	"WHERE parser = $1 AND baseline = $2 AND doc LIKE $3 " +
	"GROUP BY doc, parser, status, digest, stderr, stdout, callgrind, cfg, cfg_image " +
	"ORDER BY doc, status"

// ReportEntry an entry in the json report
type ReportEntry struct {
	Testfile           string `json:"testfile"`
	Status             string `json:"status"`
	Notes              string `json:"notes"`
	ErrorDetail        string `json:"errordetails"`
	Digest             string `json:"digest"`
	ParserFailureCount int    `json:"-"`
	ParserValidCount   int    `json:"-"`
}

// Result is the report result
type Result struct {
	Filename  string
	URL       string `db:"doc"`
	Parser    string
	Status    string
	Stderr    string
	Stdout    string
	Callgrind string
	Cfg       string
	CfgImage  []byte `db:"cfg_image"`
	Digest    string
}
