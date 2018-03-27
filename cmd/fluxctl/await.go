package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/weaveworks/flux/api"
	"github.com/weaveworks/flux/job"
	"github.com/weaveworks/flux/update"
)

var ErrTimeout = errors.New("timeout")

// await polls for a job to complete, then for the resulting commit to
// be applied
func await(ctx context.Context, stdout, stderr io.Writer, client api.Server, jobID job.ID, apply bool, verbosity int) error {
	result, err := awaitJob(ctx, client, jobID)
	if err != nil {
		return err
	}
	if result.Result != nil {
		update.PrintResults(stdout, result.Result, verbosity)
	}
	if result.Revision != "" {
		fmt.Fprintf(stderr, "Commit pushed:\t%s\n", result.Revision[:7])
	}
	if result.Result == nil {
		fmt.Fprintf(stderr, "Nothing to do\n")
		return nil
	}

	if apply && result.Revision != "" {
		if err := awaitSync(ctx, client, result.Revision); err != nil {
			return err
		}

		fmt.Fprintf(stderr, "Commit applied:\t%s\n", result.Revision[:7])
	}

	return nil
}

// await polls for a job to have been completed, with exponential backoff.
func awaitJob(ctx context.Context, client api.Server, jobID job.ID) (job.Result, error) {
	var result job.Result
	err := backoff(100*time.Millisecond, 2, 50, 1*time.Minute, func() (bool, error) {
		j, err := client.JobStatus(ctx, jobID)
		if err != nil {
			return false, err
		}
		switch j.StatusString {
		case job.StatusFailed:
			return false, j
		case job.StatusSucceeded:
			if j.Err != "" {
				// How did we succeed but still get an error!?
				return false, j
			}
			result = j.Result
			return true, nil
		}
		return false, nil
	})
	return result, err
}

// await polls for a commit to have been applied, with exponential backoff.
func awaitSync(ctx context.Context, client api.Server, revision string) error {
	return backoff(1*time.Second, 2, 10, 1*time.Minute, func() (bool, error) {
		refs, err := client.SyncStatus(ctx, revision)
		return err == nil && len(refs) == 0, err
	})
}

// backoff polls for f() to have been completed, with exponential backoff.
func backoff(initialDelay, factor, maxFactor, timeout time.Duration, f func() (bool, error)) error {
	maxDelay := initialDelay * maxFactor
	finish := time.Now().Add(timeout)
	for delay := initialDelay; time.Now().Before(finish); delay = min(delay*factor, maxDelay) {
		ok, err := f()
		if ok || err != nil {
			return err
		}
		// If we don't have time to try again, stop
		if time.Now().Add(delay).After(finish) {
			break
		}
		time.Sleep(delay)
	}
	return ErrTimeout
}

func min(t1, t2 time.Duration) time.Duration {
	if t1 < t2 {
		return t1
	}
	return t2
}
