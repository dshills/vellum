package smoke

// StartScenario bounds its retained result to one outcome. Done closes only
// after both the scenario and retirement return. Callers must wait for Done
// before reading mutable controller state; a result alone is not that join.
// Hooks are cooperative and may not be forcibly preempted by this helper.
func StartScenario(run func() error, retire func()) (result <-chan error, done <-chan struct{}) {
	outcome := make(chan error, 1)
	completed := make(chan struct{})
	go func() {
		defer close(completed)
		err := run()
		retire()
		outcome <- err
	}()
	return outcome, completed
}
