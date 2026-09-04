package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

const version = "1.0.0"

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "cron-parser %s — parse cron expressions and list upcoming runs\n\n", version)
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  cron-parser next \"<expr>\" [count]      list upcoming runs (from now)\n")
		fmt.Fprintf(os.Stderr, "  cron-parser next \"<expr>\" --from <RFC3339> [count]\n")
		fmt.Fprintf(os.Stderr, "  cron-parser matches \"<expr>\" <RFC3339> does the time match?\n")
		fmt.Fprintf(os.Stderr, "  cron-parser explain \"<expr>\"           describe each field\n")
		fmt.Fprintf(os.Stderr, "\nFormat: \"minute hour day month weekday\" — supports *, lists, ranges, steps, names (mon, jan)\n")
	}

	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	switch args[0] {
	case "next":
		if len(args) < 2 {
			usageExit("next needs an expression")
		}
		c, err := Parse(args[1])
		if err != nil {
			fail(err)
		}
		from := time.Now()
		count := 5
		rest := args[2:]
		if len(rest) > 0 && rest[0] == "--from" {
			if len(rest) < 2 {
				usageExit("--from needs an RFC3339 timestamp")
			}
			from, err = time.Parse(time.RFC3339, rest[1])
			if err != nil {
				fail(err)
			}
			rest = rest[2:]
		}
		if len(rest) > 0 {
			if _, err := fmt.Sscanf(rest[0], "%d", &count); err != nil || count < 1 || count > 100 {
				usageExit("count must be 1-100")
			}
		}
		for _, t := range c.NextN(from, count) {
			fmt.Println(t.Format(time.RFC3339))
		}

	case "matches":
		if len(args) < 3 {
			usageExit("matches needs an expression and a timestamp")
		}
		c, err := Parse(args[1])
		if err != nil {
			fail(err)
		}
		t, err := time.Parse(time.RFC3339, args[2])
		if err != nil {
			fail(err)
		}
		if c.Matches(t) {
			fmt.Println("yes")
		} else {
			fmt.Println("no")
			os.Exit(1)
		}

	case "explain":
		if len(args) < 2 {
			usageExit("explain needs an expression")
		}
		c, err := Parse(args[1])
		if err != nil {
			fail(err)
		}
		fmt.Printf("expression: %s\n", c.Original)
		fmt.Printf("  minute:  %s\n", describe(c.minute))
		fmt.Printf("  hour:    %s\n", describe(c.hour))
		fmt.Printf("  day:     %s\n", describe(c.day))
		fmt.Printf("  month:   %s\n", describe(c.month))
		fmt.Printf("  weekday: %s\n", describe(c.weekday))

	case "version":
		fmt.Println("cron-parser", version)

	default:
		flag.Usage()
		os.Exit(2)
	}
}

func describe(f field) string {
	out := ""
	first := true
	for v := f.min; v <= f.max; v++ {
		if f.values[v] {
			if !first {
				out += " "
			}
			out += fmt.Sprintf("%d", v)
			first = false
		}
	}
	return out
}

func usageExit(msg string) {
	fmt.Fprintln(os.Stderr, "error:", msg)
	os.Exit(2)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
