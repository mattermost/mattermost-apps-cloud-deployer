// Package exechelper streamlines the running of external commands while both capturing and logging
// their output.
//
// It builds on os/exec, expecting an instance of Cmd to manipulate.
package exechelper

import (
	"bufio"
	"bytes"
	"encoding/base32"
	"io"
	"math/rand"
	"os/exec"
	"sync"
	"time"

	appsutils "github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
)

// OutputLogger allows custom logging of the run command output.
type OutputLogger func(line string, logger appsutils.Logger)

var encoding = base32.NewEncoding("ybndrfg8ejkmcpqxot1uwisza345h769")
var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

// NewID is a globally unique identifier.  It is a [A-Z0-9] string 26
// characters long.  It is a UUID version 4 Guid that is zbased32 encoded
// with the padding stripped off.
func NewID() string {
	var b bytes.Buffer
	encoder := base32.NewEncoder(encoding, &b)
	encoder.Write(uuid.NewRandom())
	encoder.Close()
	b.Truncate(26) // removes the '==' padding
	return b.String()
}

func bufferAndLog(reader io.Reader, buffer *bytes.Buffer, logger appsutils.Logger, outputLogger OutputLogger) error {
	scanner := bufio.NewScanner(io.TeeReader(reader, buffer))
	for scanner.Scan() {
		text := scanner.Text()
		if outputLogger == nil {
			logger.Infof(scanner.Text())
		} else {
			outputLogger(text, logger)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

// Run invokes cmd.Run, both logging and returning STDOUT and STDERR, optionally transforming the output first.
func Run(cmd *exec.Cmd, logger appsutils.Logger, outputLogger OutputLogger) ([]byte, []byte, error) {
	// Generate a unique identifier for the command invocation by which to group logs.
	runID := NewID()

	logger = logger.With(
		"run", runID,
	)

	logger.With(
		"cmd", cmd.Path,
		"args", cmd.Args,
	).Infof("Invoking command")

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	rStdout, wStdout := io.Pipe()
	rStderr, wStderr := io.Pipe()

	cmd.Stdout = wStdout
	cmd.Stderr = wStderr

	var wg sync.WaitGroup

	// Log and buffer stdout.
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := bufferAndLog(rStdout, stdout, logger, outputLogger); err != nil {
			logger.WithError(err).Errorf("failed to scan stdout")
		}
	}()

	// Log and buffer stderr.
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := bufferAndLog(rStderr, stderr, logger, outputLogger); err != nil {
			logger.WithError(err).Errorf("failed to scan stderr")
		}
	}()

	var err error
	go func() {
		err = cmd.Run()
		wStdout.Close()
		wStderr.Close()
	}()

	wg.Wait()

	if err != nil {
		logger.WithError(err).Errorf("failed invocation")

		return stdout.Bytes(), stderr.Bytes(), errors.Wrap(err, "failed invocation")
	}

	return stdout.Bytes(), stderr.Bytes(), nil
}
