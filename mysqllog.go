// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package log4go

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
)

// This log writer sends output to a socket
type MysqlLogWriter chan *LogRecord

// This is the MysqlLogWriter's output method
func (w MysqlLogWriter) LogWrite(rec *LogRecord) {
	w <- rec
}

func (w MysqlLogWriter) Close() {
	close(w)
}

func NewMysqlLogWriter(dbName, tableName, serverId string) MysqlLogWriter {

	if serverId == "" {
		serverId = getServerId()
	}

	db, err := sql.Open("mysql", dbName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "NewMysqlLogWriter(%q,%q,%q): %s\n", dbName, tableName, serverId, err)
		return nil
	}

	w := MysqlLogWriter(make(chan *LogRecord, LogBufferLength))

	go func() {
		for rec := range w {
			// Marshall into JSON
			_, err := db.Exec("INSERT INTO "+tableName+" (ServerId, ProcessName, LogTime, LogLevel, Source, Message) VALUES (?, ?, ?, ?, ?, ?)",
				serverId, os.Args[0], rec.Created.Unix(), rec.Level.String(), rec.Source, rec.Message)
			if err != nil {
				fmt.Fprint(os.Stderr, "MysqlLogWriter(%q,%q,%q): %s", dbName, tableName, serverId, err)
				return
			}
		}
	}()

	return w
}

func getHostname() string {

	if os.Getenv("CIRCLECI") != "" {
		return ""
	}

	rsp, err := http.DefaultClient.Get("http://169.254.169.254/latest/meta-data/public-ipv4")
	if err != nil || rsp.StatusCode != http.StatusOK || rsp.Body == nil {
		return ""
	}

	defer rsp.Body.Close()
	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, rsp.Body)
	if err != nil {
		return ""
	}

	return buf.String()
}

func getServerId() string {

	return fmt.Sprintf("%s:%d", getHostname(), os.Getpid())

}
