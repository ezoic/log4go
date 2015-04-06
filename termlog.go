// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package log4go

import (
	"fmt"
	"io"
	"os"
	"time"
)

var stdout io.Writer = os.Stdout

// This is the standard writer that prints to standard output.
type ConsoleLogWriter struct {
	rec chan *LogRecord

	// The logging format
	format string
}

// This creates a new ConsoleLogWriter
func NewConsoleLogWriter() *ConsoleLogWriter {

	records := &ConsoleLogWriter{
		rec:    make(chan *LogRecord, LogBufferLength),
		format: "[%D %T] [%L] (%S) %M",
	}

	go records.run(stdout)
	return records
}

func (w *ConsoleLogWriter) run(out io.Writer) {
	for rec := range w.rec {
		fmt.Fprint(out, FormatLogRecord(w.format, rec))
		//fmt.Fprint(out, "[", timestr, "] [", levelStrings[rec.Level], "] (", rec.Source, ") ", rec.Message, "\n")
	}
}

// Set the logging format (chainable).  Must be called before the first log
// message is written.
func (w *ConsoleLogWriter) SetFormat(format string) *ConsoleLogWriter {
	w.format = format
	return w
}

// This is the ConsoleLogWriter's output method.  This will block if the output
// buffer is full.
func (w *ConsoleLogWriter) LogWrite(rec *LogRecord) {
	w.rec <- rec
}

// Close stops the logger from sending messages to standard output.  Attempts to
// send log messages to this logger after a Close have undefined behavior.
func (w *ConsoleLogWriter) Close() {
	close(w.rec)
	time.Sleep(50 * time.Millisecond) // Try to give console I/O time to complete
}
