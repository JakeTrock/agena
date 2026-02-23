package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

func main() {
	// Define flags
	listFlag := flag.Bool("list", false, "List available tasks")
	limitFlag := flag.Int("limit", 0, "Maximum number of iterations (0 = unlimited)")
	timeLimitFlag := flag.Duration("time-limit", 0*time.Second, "Maximum duration (e.g. 1h30m, 30m, 5s) (0 = unlimited)")
	taskTimeoutFlag := flag.Duration("task-timeout", 0*time.Second, "Per-candidate timeout (e.g. 5m, 30s) (overrides task.yaml)")
	geminiCommandFlag := flag.String("gemini-command", "", "Gemini command to use (overrides task.yaml)")
	dryRunFlag := flag.Bool("dry-run", false, "Print prompt without executing Gemini")
	noCleanFlag := flag.Bool("no-clean", false, "Skip startup/failure cleanup and keep workspace changes")
	verboseFlag := flag.Bool("verbose", false, "Print verbose output")
	shardFlag := flag.String("shard", "", "Shard index/total (e.g. 1/4 for first of 4 workers)")
	initFlag := flag.Bool("init", false, "Create a starter agena/ directory in the current project")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: agena <task> [options]\n")
		fmt.Fprintf(os.Stderr, "       agena --list\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	// Reorder args so flags can appear after positional args
	args := reorderArgs(os.Args[1:])
	flag.CommandLine.Parse(args)

	// Handle --init without requiring an existing agena/ directory.
	if *initFlag {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintln(os.Stderr, ColorError(fmt.Sprintf("Error: %s", formatUserError(err))))
			os.Exit(1)
		}
		created, err := InitializeAgena(cwd)
		if err != nil {
			fmt.Fprintln(os.Stderr, ColorError(fmt.Sprintf("Error: %s", formatUserError(err))))
			os.Exit(1)
		}

		if len(created) == 0 {
			fmt.Println(ColorInfo("agena/ is already initialized."))
		} else {
			fmt.Println(ColorSuccess("Created starter agena/ setup:"))
			for _, path := range created {
				fmt.Printf("  %s\n", path)
			}
		}

		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Println("  1. Review utils.py - it has TODO comments that agena will fix")
		fmt.Println("  2. Run: agena --list              (see available tasks)")
		fmt.Println("  3. Run: agena fix-todos --dry-run (preview the prompt)")
		fmt.Println("  4. Run: agena fix-todos           (fix all TODOs iteratively)")
		return
	}

	// Discover environment
	env, err := DiscoverEnvironment()
	if err != nil {
		fmt.Fprintln(os.Stderr, ColorError(fmt.Sprintf("Error: %s", formatUserError(err))))
		os.Exit(1)
	}

	// Handle --list
	if *listFlag {
		listTasks(env)
		return
	}

	// Get task name from positional args
	remaining := flag.Args()
	if len(remaining) == 0 {
		fmt.Fprintln(os.Stderr, ColorError("Error: task name required"))
		fmt.Fprintln(os.Stderr, "Use --list to see available tasks")
		os.Exit(1)
	}

	taskName := remaining[0]

	// Parse and validate shard flag (1-based indexing: 1/N through N/N)
	var partition HashPartition = NoFilter()
	if *shardFlag != "" {
		parts := strings.Split(*shardFlag, "/")
		if len(parts) != 2 {
			fmt.Fprintln(os.Stderr, ColorError("Error: --shard must be in format INDEX/TOTAL (e.g. 1/4)"))
			os.Exit(1)
		}
		index, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
		total, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err1 != nil || err2 != nil || total < 1 || index < 1 || index > total {
			fmt.Fprintln(os.Stderr, ColorError("Error: invalid shard values"))
			os.Exit(1)
		}
		partition = HashPartition{WorkerCount: total, WorkerIndex: index - 1} // Convert to 0-based internally
	}

	// Create and run the runner
	opts := RunnerOptions{
		Limit:         *limitFlag,
		TimeLimit:     *timeLimitFlag,
		DryRun:        *dryRunFlag,
		NoClean:       *noCleanFlag,
		Verbose:       *verboseFlag,
		Partition:     partition,
		Timeout:       *taskTimeoutFlag,
		GeminiCommand: *geminiCommandFlag,
	}

	runner, err := NewRunner(env, taskName, opts)
	if err != nil {
		fmt.Fprintln(os.Stderr, ColorError(fmt.Sprintf("Error: %s", formatUserError(err))))
		os.Exit(1)
	}

	if err := runner.Run(); err != nil {
		fmt.Fprintln(os.Stderr, ColorError(fmt.Sprintf("Error: %s", formatUserError(err))))
		os.Exit(1)
	}
}

func listTasks(env *Environment) {
	if len(env.Tasks) == 0 {
		fmt.Println("No tasks found.")
		return
	}

	fmt.Println(ColorBold("Available tasks:"))

	// Sort task names for consistent output
	names := make([]string, 0, len(env.Tasks))
	for name := range env.Tasks {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		task := env.Tasks[name]
		mode := "standard"
		if task.AcceptBestEffort {
			mode = "best-effort"
		}
		fmt.Printf("  %s [%s]\n", ColorInfo(fmt.Sprintf("%-30s", name)), mode)
	}
}

// reorderArgs moves flags before positional arguments so Go's flag package can parse them.
func reorderArgs(args []string) []string {
	var flags, positional []string

	i := 0
	for i < len(args) {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			flags = append(flags, arg)
			// Check if this flag takes a value (like -limit 5)
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				// Check if it's a flag that takes a value
				switch arg {
				case "-limit", "--limit", "-time-limit", "--time-limit",
					"-task-timeout", "--task-timeout", "-gemini-command", "--gemini-command",
					"-shard", "--shard":
					i++
					flags = append(flags, args[i])
				}
			}
		} else {
			positional = append(positional, arg)
		}
		i++
	}

	return append(flags, positional...)
}
