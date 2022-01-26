package cmd

import (
	"compress/gzip"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var processCmd = &cobra.Command{
	Use:   "process",
	Short: "run a process component ",
	Long:  `This subcommand runs a process component`,
	Run:   runProcessCmd,
}

func init() {
	rootCmd.AddCommand(processCmd)
	processCmd.Flags().String("parser", "", "parser to use for processing")
	processCmd.MarkFlagRequired("parser")
	processCmd.Flags().String("subset", "", "document subset to process")
	processCmd.Flags().String("tag", "", "docker tag to run")
	processCmd.MarkFlagRequired("tag")
	processCmd.Flags().String("component", "", "component to run")
	processCmd.MarkFlagRequired("component")
	processCmd.Flags().String("universe", "", "mark the processing with a universe tag")
	processCmd.MarkFlagRequired("universe")
	processCmd.Flags().Bool("baseline", false, "consider results as part of baseline")
	processCmd.Flags().Int("processMax", -1, "only process a certain number at a time")
	processCmd.Flags().Int("workerCount", 8, "number of parallel workers")
}

var port atomic.Value

func startFileServer(dir string) {
	listener, err := net.Listen("tcp4", "0.0.0.0:0")
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/", http.FileServer(http.Dir(dir)))
	log.Printf("Serving directory %s\n", dir)
	log.Println("Using port:", listener.Addr().(*net.TCPAddr).Port)
	port.Store(listener.Addr().(*net.TCPAddr).Port)
	err = http.Serve(listener, nil)
	log.Println(err)
}

func runProcessCmd(cmd *cobra.Command, args []string) {
	postgresConn := viper.Get("postgresConn").(string)
	postgresConnHost := viper.Get("postgresConnHost").(string)
	docsURL := viper.Get("docsURL").(string)
	parser, _ := cmd.Flags().GetString("parser")
	subset, _ := cmd.Flags().GetString("subset")
	component, _ := cmd.Flags().GetString("component")
	baseline, _ := cmd.Flags().GetBool("baseline")
	processMax, _ := cmd.Flags().GetInt("processMax")
	tag, _ := cmd.Flags().GetString("tag")
	workerCount, _ := cmd.Flags().GetInt("workerCount")
	universe, _ := cmd.Flags().GetString("universe")

	go startFileServer("./localdocs")
	time.Sleep(1 * time.Second)

	conn, err := pgx.Connect(context.Background(), postgresConnHost)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(context.Background())

	jobCount := 0
	jobsAlreadyProcessed := 0

	var docs []string
	// var filenames []string
	docIndexURL := fmt.Sprintf("http://127.0.0.1:%d/sd_index.gz", port.Load().(int))
	resp, err := http.Get(docIndexURL)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	fz, err := gzip.NewReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	defer fz.Close()
	body, err := ioutil.ReadAll(fz)
	if err != nil {
		log.Fatal(err)
	}

	var m map[string]int = make(map[string]int)

	var existsParserColumn = "SELECT column_name from information_schema.columns WHERE table_name = '" + strings.ReplaceAll(component, "-", "_") + "' AND column_name = 'parser'"
	var columnName string
	result := conn.QueryRow(context.Background(), existsParserColumn).Scan(&columnName)
	parserColumnExists := true
	if result != nil {
		parserColumnExists = false
	}
	var existsToolRunQuery string
	var rows pgx.Rows
	if parserColumnExists {
		existsToolRunQuery = "SELECT substring(doc from '(?:.+/)(.+)') AS filename FROM " + strings.ReplaceAll(component, "-", "_") + " WHERE parser = $1 AND baseline = $2"
		rows, err = conn.Query(context.Background(), existsToolRunQuery, parser, baseline)
	} else {
		existsToolRunQuery = "SELECT substring(doc from '(?:.+/)(.+)') AS filename FROM " + strings.ReplaceAll(component, "-", "_") + " WHERE baseline = $1"
		rows, err = conn.Query(context.Background(), existsToolRunQuery, baseline)
	}
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			log.Fatalf("cannot read row: %v", err)
		}
		fname := url[strings.LastIndex(url, "/")+1:]
		m[fname] = 1
	}

	for _, line := range strings.Split(string(body), "\n") {
		if strings.Contains(line, subset) {
			filename := line[strings.LastIndex(line, "/")+1:]
			filename = strings.Replace(filename, " ", "%20", -1)
			_, exists := m[filename]
			if !exists {
				docs = append(docs, strings.Replace(line, " ", "%20", -1))
			}
		}
	}

	numJobs := len(docs)
	jobs := make(chan BatchJob, numJobs)
	results := make(chan string, numJobs)
	for w := 0; w < workerCount; w++ {
		go worker(w, jobs, results)
	}

	containerBatchSize := 10
	for i := 0; i < len(docs); i += containerBatchSize {
		if processMax > -1 && jobCount >= processMax*containerBatchSize {
			log.Println("processMax: submitted maximum jobs, waiting for completion...")
			break
		}

		batchJob := BatchJob{}

		for j := 0; j < containerBatchSize; j++ {
			if len(docs) > i+j {
				mrDocURL := fmt.Sprintf("http://%s:%d/%s", docsURL, port.Load().(int), docs[i+j])
				batchJob.Meta.DocURL += mrDocURL + " "
			}
		}

		batchJob.Tag = tag
		batchJob.Meta.PostgresConn = postgresConn
		batchJob.Meta.IsBaseline = strconv.FormatBool(baseline)
		batchJob.Meta.Universe = universe
		batchJob.Meta.Parser = parser

		jobs <- batchJob
		jobCount++
	}

	finishedJobs := 0
	if jobCount > 0 {
		for {
			<-results
			finishedJobs++
			log.Println(finishedJobs, "containers completed")
			if finishedJobs >= jobCount {
				// os.Exit(0)
				break
			}
		}
	} else {
		log.Println("no jobs were queued")
		log.Println(jobsAlreadyProcessed, "jobs were not started because they have already been run to completion")
	}
}

// BatchJob batch job spec
type BatchJob struct {
	Payload string
	Tag     string
	Meta    Meta
}

// Meta metadata for job
type Meta struct {
	Parser       string `json:"MR_PARSER"`
	IsBaseline   string `json:"MR_IS_BASELINE"`
	DocURL       string `json:"MR_DOC_URL"`
	PostgresConn string `json:"MR_POSTGRES_CONN"`
	Universe     string `json:"MR_UNIVERSE"`
}

func worker(id int, jobs <-chan BatchJob, results chan<- string) {
	for j := range jobs {
		log.Println("worker: " + strconv.Itoa(id) + " took job for " + j.Meta.DocURL)
		cmd := exec.Command("docker", "run", "--add-host=host.docker.internal:host-gateway", "--rm",
			"-e", "MR_DOC_URL="+j.Meta.DocURL, "-e", "MR_POSTGRES_CONN="+j.Meta.PostgresConn,
			"-e", "MR_PARSER="+j.Meta.Parser, "-e", "MR_IS_BASELINE="+j.Meta.IsBaseline,
			"-e", "MR_UNIVERSE="+j.Meta.Universe,
			j.Tag)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			log.Println(err)
		}

		results <- j.Meta.DocURL
	}
}
