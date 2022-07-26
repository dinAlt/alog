package main

import (
	"alog/alog"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"sync"
	"time"
)

var (
	capacity   int
	help       bool
	numProcs   int
	numRecords int
	benchStd   bool
)

func init() {
	flag.IntVar(&capacity, "capacity", alog.DefaultCapacity, "capacity of log channel")
	flag.BoolVar(&help, "h", false, "print this help message")
	flag.BoolVar(&benchStd, "bench-std", false, "also bench std log")
}

func main() {
	parseCmd()

	reportBuff := &bytes.Buffer{}

	benchCustomLogger(reportBuff)
	if benchStd {
		benchStdLogger(reportBuff)
	}

	fmt.Print(reportBuff)
}

// benchCustomLogger runs benchmark for alog.Logger
func benchCustomLogger(w io.Writer) time.Duration {
	l := alog.New(alog.WithCapacity(capacity))
	var start, end time.Time

	defer func() {
		err := l.Stop(context.TODO())
		if err != nil {
			panic(err)
		}
		_, _ = fmt.Fprintf(w, "benchCustomLogger: goroutines done in: %v\n", end.Sub(start))
		_, _ = fmt.Fprintf(w, "benchCustomLogger: done in: %v\n", time.Since(start))
	}()

	go func() {
		err := l.Run()
		if err != nil {
			panic(err)
		}
	}()

	wg := sync.WaitGroup{}
	start = time.Now()

	wg.Add(numProcs)
	for procNo := 0; procNo < numProcs; procNo++ {
		go func(procNo int) {
			for recNo := 0; recNo < numRecords; recNo++ {
				rec := recNo
				l.Printf("Goroutine %d: message %d", procNo, rec)
			}
			wg.Done()
		}(procNo)
	}
	wg.Wait()

	end = time.Now()
	return end.Sub(start)
}

// benchCustomLogger runs benchmark for log package
func benchStdLogger(w io.Writer) time.Duration {
	wg := sync.WaitGroup{}
	start := time.Now()

	wg.Add(numProcs)
	for procNo := 0; procNo < numProcs; procNo++ {
		go func(procNo int) {
			for recNo := 0; recNo < numRecords; recNo++ {
				rec := recNo
				log.Printf("Goroutine %d: message %d", procNo, rec)
			}
			wg.Done()
		}(procNo)
	}
	wg.Wait()

	_, _ = fmt.Fprintf(w, "benchStdLogger: Std log: done in: %v\n", time.Since(start))
	return time.Until(start)
}

func parseCmd() {
	flag.Parse()
	if help || flag.NArg() < 2 {
		printUsage()
		os.Exit(0)
	}
	var err error
	numProcs, err = strconv.Atoi(flag.Arg(0))
	if err != nil {
		fmt.Printf("%v", err)
		printUsage()
		os.Exit(1)
	}
	numRecords, err = strconv.Atoi(flag.Arg(1))
	if err != nil {
		fmt.Printf("%v", err)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Printf("  %s <NUM_PROCS> <NUM_RECORDS>\n\n", os.Args[0])
	fmt.Println("Flags:")
	flag.PrintDefaults()
}
