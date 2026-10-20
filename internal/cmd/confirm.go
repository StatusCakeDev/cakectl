package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"strings"
)

// ErrNotConfirmed reports a destructive operation the user declined.
var ErrNotConfirmed = errors.New("aborted: not confirmed")

// confirm asks the user to approve a destructive operation.
//
// With yes set it approves without asking. Otherwise it asks on the terminal,
// and refuses outright when input is not a terminal, because a script must not
// delete things on the strength of an answer nobody was there to give.
func (o *RootOptions) confirm(prompt string, yes bool) error {
	if yes {
		return nil
	}

	if !o.isTerminal(o.In) {
		return errors.New("refusing to continue without confirmation: input is not a terminal; pass --yes to confirm")
	}

	fmt.Fprintf(o.ErrOut, "%s [y/N]: ", prompt)

	answer, err := bufio.NewReader(o.In).ReadString('\n')
	if err != nil && answer == "" {
		return ErrNotConfirmed
	}

	switch strings.ToLower(strings.TrimSpace(answer)) {
	case "y", "yes":
		return nil
	default:
		return ErrNotConfirmed
	}
}
