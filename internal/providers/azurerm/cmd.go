package azurerm

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var defaultAzBinary = "az"

type CmdOptions struct {
	Binary string
	Flags  []string
	Dir    string
}

type CmdError struct {
	err    error
	Stderr []byte
}

func (e *CmdError) Error() string {
	return e.err.Error()
}

type cmdLogger interface {
	Log(level log.Level, args ...interface{})
}

func Cmd(opts *CmdOptions, args ...string) ([]byte, error) {
	exe := opts.Binary
	if exe == "" {
		exe = defaultAzBinary
	}

	cmd := exec.Command(exe, append(args, opts.Flags...)...)
	log.Debugf("Running command: %s", cmd.String())
	cmd.Dir = opts.Dir
	cmd.Env = os.Environ()

	// TODO: Check necessity of additional ENV variables below

	logWriter := &cmdLogWriter{
		logger: log.StandardLogger().WithField("binary", "az"),
		level:  log.DebugLevel,
	}
	azLogWriter := &cmdLogWriter{
		logger: log.StandardLogger().WithField("binary", "az"),
		level:  log.DebugLevel,
	}

	var outbuf bytes.Buffer
	outw := bufio.NewWriter(&outbuf)
	var errbuf bytes.Buffer
	errw := bufio.NewWriter(&errbuf)

	cmd.Stdout = io.MultiWriter(outw, azLogWriter)
	cmd.Stderr = io.MultiWriter(errw, logWriter)
	err := cmd.Run()

	outw.Flush()
	errw.Flush()
	azLogWriter.Flush()
	logWriter.Flush()

	if err != nil {
		return outbuf.Bytes(), &CmdError{err, errbuf.Bytes()}
	}

	return outbuf.Bytes(), nil
}

func getArmTemplateFromBicepProject() {
	panic("Not implemented")
}

func getWhatIfFromArmTemplate(templateFile string, opts *ArmTemplateProviderOpts) ([]byte, error) {
	var args []string
	var err error

	switch opts.Scope {
	case ResourceGroup:
		args, err = getGroupDeploymentArgs(templateFile, opts)
	default:
		err = errors.New(fmt.Sprintf("Unsupported scope %s", opts.Scope))
	}
	if err != nil {
		return nil, err
	}

	templateDir := filepath.Dir(templateFile)
	cmdOpts := &CmdOptions{
		Binary: opts.Binary,
		Flags:  args,
		Dir:    templateDir,
	}

	output, err := Cmd(cmdOpts)
	if err != nil {
		return nil, err
	}

	return output, nil
}

// Adapted from https://github.com/sirupsen/logrus/issues/564#issuecomment-345471558
// Needed to ensure we can log large Terraform output lines.
type cmdLogWriter struct {
	logger cmdLogger
	level  log.Level
	buf    bytes.Buffer
	mu     sync.Mutex
}

func (w *cmdLogWriter) Write(b []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	origLen := len(b)
	for {
		if len(b) == 0 {
			return origLen, nil
		}
		i := bytes.IndexByte(b, '\n')
		if i < 0 {
			w.buf.Write(b)
			return origLen, nil
		}

		w.buf.Write(b[:i])
		w.alwaysFlush()
		b = b[i+1:]
	}
}

func (w *cmdLogWriter) alwaysFlush() {
	w.logger.Log(w.level, w.buf.String())
	w.buf.Reset()
}

func (w *cmdLogWriter) Flush() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.buf.Len() != 0 {
		w.alwaysFlush()
	}
}
