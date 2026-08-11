// `gunzip -c suneidotypes.log.gz`
package tlog

import (
	"compress/gzip"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/apmckinlay/gsuneido/util/assert"
)

// compressed log file size threshold in megabytes
const maxSizeMB = 75

// 0 = truncate current log file
const maxBackups = 2

const logFile = "suneidotypes.log.gz"

var (
	mu  sync.Mutex
	out *os.File
	gz  *gzip.Writer
)

// it is fine to decode concatenaded gzip files
func Setup() {
	f, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tlog: cannot open %s: %v\n", logFile, err)
		return
	}
	out = f
	gz = gzip.NewWriter(f)
}

// safe to call when logging is disables
func Close() {
	mu.Lock()
	defer mu.Unlock()
	if gz != nil {
		_ = gz.Close()
		gz = nil
	}
	if out != nil {
		_ = out.Close()
		out = nil
	}
}

// Enabled reports whether logging is set up. Logf is already a no-op when it
// is not, so this is only for callers that would otherwise pay to build an
// expensive argument - see Process, which would marshal the whole request.
func Enabled() bool {
	mu.Lock()
	defer mu.Unlock()
	return gz != nil
}

func Logf(format string, args ...any) {
	mu.Lock()
	defer mu.Unlock()
	if gz == nil {
		return
	}
	enforceSizeLimit()
	if gz == nil {
		return // rotate() failed to reopen; logging is now disabled
	}
	_, _ = fmt.Fprintf(gz, "%s %s\n", time.Now().Format("2006-01-02 15:04:05.000"),
		fmt.Sprintf(format, args...))
	_ = gz.Flush()
}

// Caller must hold mutex
func enforceSizeLimit() {
	assert.That(maxSizeMB > 0)
	info, err := out.Stat()
	if err != nil {
		return
	}
	if info.Size() < int64(maxSizeMB)*1024*1024 {
		return
	}
	rotate()
}

func rotate() {
	_ = gz.Close()
	_ = out.Close()
	gz, out = nil, nil

	if maxBackups <= 0 {
		// No backups: drop the file entirely and start over.
		_ = os.Remove(logFile)
	} else {
		_ = os.Remove(backupName(maxBackups))
		for i := maxBackups - 1; i >= 1; i-- {
			// Ignore "not exist" errors: not every slot is populated yet.
			_ = os.Rename(backupName(i), backupName(i+1))
		}
		_ = os.Rename(logFile, backupName(1))
	}

	f, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		// Could not reopen; leave logging disabled rather than panic
		fmt.Fprintf(os.Stderr, "tlog: cannot reopen %s after rotate: %v\n", logFile, err)
		return
	}
	out = f
	gz = gzip.NewWriter(f)
}

func backupName(i int) string {
	return fmt.Sprintf("suneidotypes.log.%d.gz", i)
}
