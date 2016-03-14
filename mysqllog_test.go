package log4go

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/kylelemons/godebug/pretty"
)

func Test_MysqlLogWriter(t *testing.T) {

	testServerId := "ezoictestserverid"
	dsn := os.Getenv("MYSQL_DSN")
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec("DELETE FROM EzoicMonitor.ProcessLog WHERE ServerId = ?", testServerId)
	if err != nil {
		t.Fatal(err)
	}

	pc, _, lineno, ok := runtime.Caller(2)
	src := ""
	if ok {
		src = fmt.Sprintf("%s:%d", runtime.FuncForPC(pc).Name(), lineno)
	}

	rt := time.Unix(time.Now().Unix(), 0)

	rec := &LogRecord{
		Level:   DEBUG,
		Created: rt,
		Source:  src,
		Message: "test message",
	}

	log.Println("creating log writer")
	w := NewMysqlLogWriter(dsn, "", testServerId)
	log.Println("writing")
	w.LogWrite(rec)
	log.Println("closing")
	w.Close()

	log.Println("checking")

	var dbLogTime int64
	var dbLogLevel, dbSource, dbMessage string
	err = db.QueryRow("SELECT LogTime, LogLevel, Source, Message FROM EzoicMonitor.ProcessLog WHERE Serverid = ?", testServerId).Scan(&dbLogTime, &dbLogLevel, &dbSource, &dbMessage)
	if err != nil {
		t.Fatalf("no mysql records found, %v", err)
	}

	outRec := &LogRecord{}
	for i, ll := range levelStrings {
		if ll == dbLogLevel {
			outRec.Level = Level(i)
			break
		}
	}
	outRec.Created = time.Unix(dbLogTime, 0)
	outRec.Source = dbSource
	outRec.Message = dbMessage

	if reflect.DeepEqual(outRec, rec) == false {
		t.Fatalf("unexpected row\n%s", pretty.CompareConfig.Compare(outRec, rec))
	}

}
