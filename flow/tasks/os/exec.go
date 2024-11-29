package os

import (
	"bytes"
	"time"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"

	"cuelang.org/go/cue"
	hofcontext "github.com/hofstadter-io/hof/flow/context"
	"github.com/hofstadter-io/hof/lib/cuetils"
)

type Exec struct {
	cmd    *exec.Cmd
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewExec(val cue.Value) (hofcontext.Runner, error) {
	return &Exec{}, nil
}

func (T *Exec) Run(hctx *hofcontext.Context) (interface{}, error) {
	
	v := hctx.Value

	// Create a cancellable context for process management
	goctx, cancel := context.WithCancel(context.Background())
	T.cancel = cancel
	// Add defer at start of Run method
	defer func() {
			if T.cancel != nil {
					T.cancel()
			}
	}()

	// Setup channels for process management
	processComplete := make(chan struct{}, 1)
	processError := make(chan error, 1)

	// Setup return value with defaults
	ret := map[string]interface{}{
		"exitcode": -1,
		"success":  false,
	}

	// Lock CUE access and extract config
	init_schemas(v.Context())
	v = v.Unify(task_exec)
	if v.Err() != nil {
		return nil, cuetils.ExpandCueError(v.Err())
	}
	
	// Extract all configuration under lock
	hctx.CUELock.Lock()
	cmds, dir, env, stdin, stdout, stderr, doExit, exterr := extractExecConfig(v)
	hctx.CUELock.Unlock()

	if exterr != nil {
		return nil, fmt.Errorf("os.Exec field: value extraction failed: %w", exterr)
	}

	// Create and configure command with goctx
	T.cmd = exec.CommandContext(goctx, cmds[0], cmds[1:]...)
	T.cmd.Dir = dir
	T.cmd.Env = env

	// Setup I/O
	if stdin != nil {
		T.cmd.Stdin = stdin
	}
	if stdout != nil {
		T.cmd.Stdout = stdout
	}
	if stderr != nil {
		T.cmd.Stderr = stderr
	}

	// Launch process in background
	T.wg.Add(1)
	go func() {
		defer T.wg.Done()
		defer close(processComplete)
		defer close(processError)
		
		// Start the process
		if err := T.cmd.Start(); err != nil {
			processError <- err
			return
		}

		// Wait for process completion
		err := T.cmd.Wait()
		if err != nil {
			processError <- err
		}
		processComplete <- struct{}{} 
	}()

	// Wait for completion, error, or cancellation
	select {
	case <-processComplete:
		// Update return value with process state
		if T.cmd.ProcessState != nil {
			ret["exitcode"] = T.cmd.ProcessState.ExitCode()
			ret["success"] = T.cmd.ProcessState.Success()
		}
		ret, _ = fillIO(v, ret, stdout, stderr)

	case err := <-processError:
		// Handle error according to user preferences
		ret["error"] = err.Error()
		if doExit {
			return ret, err
		}

	case <-goctx.Done():
		// Flow context cancelled (CTRL-C, timeout, etc)
		T.cleanup()
		ret["error"] = "flow cancelled"

		// Capture any output that occurred before cancellation
		ret, _ = fillIO(v, ret, stdout, stderr) 

		return ret, goctx.Err()
	}

	return ret, nil
}

func (T *Exec) cleanup() {
	// Cancel our process context
	if T.cancel != nil {
		T.cancel()
	}

	// Try graceful shutdown first
	if T.cmd != nil && T.cmd.Process != nil {
		_ = T.cmd.Process.Signal(syscall.SIGTERM)
	}

	// Wait briefly for cleanup
	done := make(chan struct{})
	go func() {
		T.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return
	case <-time.After(3 * time.Second):
		// Force kill if still running
		if T.cmd != nil && T.cmd.Process != nil {
			_ = T.cmd.Process.Kill()
		}
	}
}

func extractExecConfig(v cue.Value) (
    cmds []string,
    dir string,
    env []string,
    stdin io.Reader,
    stdout io.Writer,
    stderr io.Writer,
    doExit bool,
    err error,
) {
    // Extract command
    cmds, err = extractCmd(v)
    if err != nil {
        return
    }

    // Extract directory
    dir, err = extractDir(v)
    if err != nil {
        return
    }

    // Extract environment
    env, err = extractEnv(v)
    if err != nil {
        return
    }

    // Extract I/O configuration
    stdin, stdout, stderr, err = extractIO(v)
    if err != nil {
        return
    }

    // Extract exit behavior
    doExit, err = extractExit(v)
    return
}

func extractCmd(ex cue.Value) ([]string, error) {
	val := ex.LookupPath(cue.ParsePath("cmd"))
	if val.Err() != nil {
		return nil, val.Err()
	}

	cmds := []string{}
	switch val.IncompleteKind() {
	case cue.StringKind:
		c, err := val.String()
		if err != nil {
			return nil, err
		}
		cmds = []string{c}
	case cue.ListKind:
		l, err := val.List()
		if err != nil {
			return nil, err
		}
		for l.Next() {
			c, err := l.Value().String()
			if err != nil {
				return nil, err
			}
			cmds = append(cmds, c)
		}
	default:
		return nil, fmt.Errorf("unsupported cmd type: %T", val.IncompleteKind())
	}

	return cmds, nil
}

func extractDir(ex cue.Value) (string, error) {
	// handle Stdout
	d := ex.LookupPath(cue.ParsePath("dir"))
	if d.Exists() {
		s, err := d.String()
		if err != nil {
			return "", err
		}
		return s, nil
	}
	return "", nil
}

func extractExit(ex cue.Value) (bool, error) {
	// handle Stdout
	d := ex.LookupPath(cue.ParsePath("exitonerr"))
	if d.Exists() {
		b, err := d.Bool()
		if err != nil {
			return true, err
		}
		return b, nil
	}
	return true, nil
}

func extractEnv(ex cue.Value) ([]string, error) {

	val := ex.LookupPath(cue.ParsePath("env"))
	if val.Exists() {
		// convert env map in CUE to slice in go
		env := make([]string, 0)
		iter, err := val.Fields()
		if err != nil {
			return nil, err
		}
		for iter.Next() {
			k := iter.Selector().String()
			if err != nil {
				return nil, err
			}
			v, err := iter.Value().String()
			if err != nil {
				return nil, err
			}
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
		return env, nil
	}

	return nil, nil
}

// extractIO handles configuring input/output streams for command execution based on CUE configuration.
// It supports several IO modes for each stream:
// - String input is treated as direct input (except "-" which means use stdin)
// - Bytes input is used directly
// - Boolean true connects to process standard streams
// - Null or unspecified uses sensible defaults
func extractIO(ex cue.Value) (Stdin io.Reader, Stdout, Stderr io.Writer, err error) {

		// handle stdin,
    iv := ex.LookupPath(cue.ParsePath("stdin"))
    if iv.Exists() {
        switch iv.IncompleteKind() {
        case cue.StringKind:
            s, err := iv.String()
            if err != nil {
                return nil, nil, nil, err
            }
            // Special case: "-" means use standard input
            // This allows tasks to specify they want to read from stdin
            if s == "-" {
								// (BUG): works around centralized printing
                Stdin = os.Stdin
            } else {
                // Otherwise create a reader from the string content
                Stdin = strings.NewReader(s)
            }

        case cue.BytesKind:
            b, err := iv.Bytes()
            if err != nil {
                return nil, nil, nil, err
            }
            // Create a reader directly from bytes
            Stdin = bytes.NewReader(b)

        case cue.BoolKind:
            Stdin = os.Stdin

        case cue.NullKind:
						// do nothing so no Stdin is set
        
        default:
            return nil, nil, nil, fmt.Errorf("unsupported type %v for stdin", iv.IncompleteKind())
        }
    }

    // Handle stdout configuration
    ov := ex.LookupPath(cue.ParsePath("stdout"))
    if !ov.Exists() {
        // Default to process stdout if not specified
        Stdout = os.Stdout
    } else {
        switch ov.IncompleteKind() {
				// we want a bytes writer for Now
				// will return the proper format when filling the value back
        case cue.StringKind:
            fallthrough
        case cue.BytesKind:
            Stdout = new(bytes.Buffer)

        case cue.BoolKind:
            Stdout = os.Stdout

        case cue.NullKind:
            // do nothing so no Stdout is set
        
        default:
            return nil, nil, nil, fmt.Errorf("unsupported type %v for stdout", ov.IncompleteKind())
        }
    }

    // Handle stderr configuration
    // This follows the same pattern as stdout
    ev := ex.LookupPath(cue.ParsePath("stderr"))
    if !ev.Exists() {
        // Default to process stderr if not specified
        Stderr = os.Stderr
    } else {
        switch ev.IncompleteKind() {
				// we want a bytes writer for Now
				// will return the proper format when filling the value back
        case cue.StringKind:
            fallthrough
        case cue.BytesKind:
            Stderr = new(bytes.Buffer)

        case cue.BoolKind:
            Stderr = os.Stderr

        case cue.NullKind:
            // do nothing so no Stderr is set
        
        default:
            return nil, nil, nil, fmt.Errorf("unsupported type %v for stderr", ev.IncompleteKind())
        }
    }

    return Stdin, Stdout, Stderr, nil
}

func fillIO(ex cue.Value, ret map[string]interface{}, Stdout, Stderr io.Writer) (map[string]interface{}, error) {
	// (warn) possible cue evaluator race condition here
	ov := ex.LookupPath(cue.ParsePath("stdout"))
	if ov.Exists() {
		switch ov.IncompleteKind() {
		// we want a bytes writer for Now
		// will return the proper format when filling the value back
		case cue.StringKind:
			buf := Stdout.(*bytes.Buffer)
			ret["stdout"] = buf.String()
		case cue.BytesKind:
			buf := Stdout.(*bytes.Buffer)
			ret["stdout"] = buf.Bytes()
		case cue.NullKind:
			// do nothing, Stdout was not captured
		}
	}

	ev := ex.LookupPath(cue.ParsePath("stderr"))
	if ev.Exists() {
		switch ev.IncompleteKind() {
		// we want a bytes writer for Now
		// will return the proper format when filling the value back
		case cue.StringKind:
			buf := Stderr.(*bytes.Buffer)
			ret["stderr"] = buf.String()
		case cue.BytesKind:
			buf := Stderr.(*bytes.Buffer)
			ret["stderr"] = buf.Bytes()
		case cue.NullKind:
			// do nothing, Stderr was not captured
		}
	}

	return ret, nil
}
