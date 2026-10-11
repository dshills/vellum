package replayconsumer

import (
	"context"
	"fmt"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/execution"
)

type Checkpoint struct {
	Action string
	State  Snapshot
}
type CommandProof struct {
	Kind          string
	Batch         int
	FixtureSHA256 string
	Checkpoints   []Checkpoint
	Final         Snapshot
	Mailbox       app.PendingState
	Executor      execution.Snapshot
}

// RunCommands qualifies only the command route. Native acceptance must drive
// the declared View via actual pointer/key routing and capture measured layout.
func RunCommands(ctx context.Context, batch int) (proof CommandProof, err error) {
	if ctx == nil || batch < 1 || batch > 64 {
		return proof, fmt.Errorf("consumer: context and batch 1..64 required")
	}
	if err := ctx.Err(); err != nil {
		return proof, err
	}
	a, err := app.New(app.Config{QueueCapacity: 32, CompletionCapacity: 4})
	if err != nil {
		return proof, err
	}
	defer a.Close()
	ex, err := execution.NewManual(4)
	if err != nil {
		return proof, err
	}
	defer ex.Close()
	f, err := New(Config{Owner: a, Executor: ex})
	if err != nil {
		return proof, err
	}
	proof = CommandProof{Kind: "portable-command-only", Batch: batch, FixtureSHA256: FixtureSHA256()}
	actions := []Action{{ID: OpenProject}, {ID: EditProperty, Value: "Harbor"}, {ID: SortTable}, {ID: SelectRow, Value: "0"}, {ID: StartTask}, {ID: CancelTask}, {ID: Undo}}
	var commandErr error
	for _, action := range actions {
		action := action
		if err := a.PostWindow(f.Window, func() {
			if commandErr != nil {
				return
			}
			if commandErr = f.Invoke(ctx, action); commandErr != nil {
				return
			}
			if commandErr = f.Validate(); commandErr != nil {
				return
			}
			proof.Checkpoints = append(proof.Checkpoints, Checkpoint{Action: action.ID, State: f.Snapshot()})
		}); err != nil {
			return proof, err
		}
	}
	if err := drainReady(ctx, a, batch); err != nil {
		return proof, err
	}
	if commandErr != nil {
		return proof, commandErr
	}
	if _, err := ex.StepNext(ctx); err != nil {
		return proof, err
	}
	if err := drainReady(ctx, a, batch); err != nil {
		return proof, err
	}
	if err := a.Do(func() { proof.Final = f.Snapshot(); commandErr = f.Validate() }); err != nil {
		return proof, err
	}
	if commandErr != nil {
		return proof, commandErr
	}
	proof.Mailbox, proof.Executor = a.Pending(), ex.Snapshot()
	if !proof.Mailbox.Idle() || !proof.Executor.Idle() {
		return proof, fmt.Errorf("consumer: residual command work app%+v executor%+v", proof.Mailbox, proof.Executor)
	}
	if len(proof.Final.UIAttempts) != 0 {
		return proof, fmt.Errorf("consumer: command proof invoked widget callbacks")
	}
	return proof, nil
}

func drainReady(ctx context.Context, a *app.Application, batch int) error {
	for range 128 {
		if err := ctx.Err(); err != nil {
			return err
		}
		pending := a.Pending()
		if pending.Ordinary == 0 && pending.ReadyCompletions == 0 {
			return nil
		}
		processed, err := a.DrainN(batch)
		if err != nil {
			return err
		}
		if processed == 0 {
			return fmt.Errorf("consumer: ready mailbox made no progress: %+v", pending)
		}
	}
	return fmt.Errorf("consumer: owner step budget exhausted")
}
