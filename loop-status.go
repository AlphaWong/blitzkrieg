package blast

import (
	"context"
	"fmt"
	"io"
	"time"
)

func (b *Blaster) startStatusLoop(ctx context.Context) {

	if b.Quiet || b.outWriter == nil {
		return
	}

	b.mainWait.Add(1)
	ticker := time.NewTicker(time.Second * 10)

	go func() {
		defer b.mainWait.Done()
		defer b.println("Exiting status loop")
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-b.dataFinishedChannel:
				ticker.Stop()
				return
			case <-ticker.C:
				// notest
				b.printStatus(false)
			}
		}
	}()
}

// PrintStatus prints the status message to the output writer
func (b *Blaster) PrintStatus(writer io.Writer) {
	fmt.Fprint(writer, b.Stats())
}

func (b *Blaster) printStatus(final bool) {

	if b.Quiet || b.outWriter == nil {
		return
	}

	b.PrintStatus(b.outWriter)

	if !final {
		b.printRatePrompt()
	}
}

func (b *Blaster) printRatePrompt() {

	if b.inputReader == nil {
		return
	}

	b.printf(`
Current rate is %.0f requests / second. Enter a new rate or press enter to view status.

Rate?
`,
		b.Rate,
	)
}

func (b *Blaster) println(a ...interface{}) {
	if b.Quiet || b.outWriter == nil {
		return
	}
	fmt.Fprintln(b.outWriter, a...)
}

func (b *Blaster) printf(format string, a ...interface{}) {
	if b.Quiet || b.outWriter == nil {
		return
	}
	fmt.Fprintf(b.outWriter, format, a...)
}
