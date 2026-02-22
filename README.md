# Agena

Agena is a tool for automating iterative code improvements using Gemini. You specify a task and a set of candidates. Agena works through the task with appropriate success/failure handling, logging, guardrails, etc.

## How It Works

1. Runs your **candidate source** to identify tasks to run
2. Selects the first unprocessed candidate
3. Sends candidate details to Gemini with your templated prompt
4. Runs your **verify command** (e.g. `cargo check`)
5. Checks if the candidate is still present in the source
6. Commits successful tasks or resets the project on failure
7. Repeats until done or limit reached

A task is only considered successful if the candidate disappears from the source. This can be relaxed with `accept_best_effort: true` if you want to keep partial improvements.

## Why Agena?

Agena may be a good fix for you if you need to run a large number of long-running tasks, which in turn may need to be tuned/customised over time.

I love Agena because:
* Tasks are expressed via configuration: it is to experiment with new ideas by copying an existing task and tweaking it;
* Candidate sources are just the JSON / newline delimited output of shell commands so it's easy to drop in existing scripts or write new ones. There's no special schema.
* Gemini's output is streamed and presented to you like a normal session despite you running in non-interactive mode.
* You can tell Agena to stop after the current task finishes with Ctrl-\\.
* Built in parallelism support with --shard.
* Agena is extensively tested with both unit and integration tests.

## Requirements

- Go 1.21+
- [Gemini CLI](https://github.com/google/gemini-cli) installed and authenticated

## Installation

```bash
go build -o bin/agena ./src
```

Then add the bin directory to your path:

```bash
export PATH="$PATH:/your/path/to/agena/bin"
```

## Quick Start

1. Create an `agena/` directory in your project root
2. Add a `config.yaml` with global settings
3. Create task directories with `task.yaml` files
4. Run: `agena <task-name>`

Minimal example:

```
project-root/
├── agena/
│   ├── config.yaml
│   └── fix-errors/
│       └── task.yaml
```

**config.yaml:**
```yaml
gemini_command: "gemini"
verify_command: "cargo check"
success_command: "git commit -m 'Fix: $CANDIDATE'"
reset_command: "git reset --hard"
```

**fix-errors/task.yaml:**
```yaml
candidate_source: "cargo check 2>&1 | grep -oP 'error\\[E\\d+\\].*' | jq -R -s 'split(\"\\n\") | map(select(length > 0))'",
prompt: "Fix this compiler error: $INPUT"
```

You can override the global `gemini_command` for a specific task by adding `gemini_command` to your `task.yaml`:

## Usage

```bash
# List available tasks
agena --list

# Create starter agena/ scaffolding
agena --init

# Run a task
agena mytask

# Run with iteration limit
agena mytask --limit 10

# Preview prompts without executing
agena mytask --dry-run --verbose

# Distribute work across parallel runners
agena mytask --shard 1/4  # Terminal 1 (first of 4 workers)
agena mytask --shard 2/4  # Terminal 2 (second of 4 workers)
agena mytask --shard 3/4  # Terminal 3
agena mytask --shard 4/4  # Terminal 4

# Override task settings temporarily
agena mytask --task-timeout 5m      # Per-candidate timeout
agena mytask --gemini-command "gemini"
```

If `agena/` is missing, Agena now auto-creates a starter setup in the current directory.

| Flag                | Description                                         |
| ------------------- | --------------------------------------------------- |
| `--list`            | List all available tasks                            |
| `--limit N`         | Maximum iterations (0 = unlimited)                  |
| `--time-limit`      | Maximum duration for entire task run                |
| `--task-timeout`    | Per-candidate timeout (overrides task.yaml)         |
| `--gemini-command`  | Gemini command to use (overrides task.yaml)         |
| `--init`            | Create starter `agena/` config and example task     |
| `--dry-run`         | Print prompts without executing Gemini              |
| `--verbose`         | Print full prompt content and show command overrides |
| `--shard I/N`       | Shard index/total for parallel processing           |

## Configuration

### config.yaml (Global)

```yaml
# Path to Gemini CLI
gemini_command: "gemini"

# Runs after Gemini makes changes, before checking if candidate is resolved
verify_command: "cargo check"

# Runs when candidate is no longer present in source
# Available variables: $CANDIDATE (JSON), $TASK_NAME
success_command: "git commit -m 'Fix: $CANDIDATE'"

# Runs when candidate is still present (or verify failed)
reset_command: "git reset --hard"
```

### task.yaml (Per-Task)

```yaml
candidate_source: "cargo check 2>&1 | grep error"
prompt: "Fix this issue: $INPUT"       # Inline prompt, or...
template: "template.txt"               # ...load from file
gemini_flags: ""                       # Optional CLI flags
gemini_command: "gemini"               # Override global gemini_command
accept_best_effort: false              # Accept partial fixes
timeout: "5m"                          # Per-candidate timeout (optional)
```

**Timeouts**

The `timeout` option limits how long Gemini can spend on a single candidate. When timeout is reached, Gemini is interrupted and Agena handles the current work:

- If `accept_best_effort: true` and build passes, commits partial progress
- Otherwise, resets changes and marks candidate as ignored

Duration format: `30s`, `5m`, `1h`, etc. (Go `time.ParseDuration` format).

This is different from the `--time-limit` CLI flag which applies to the entire task run. Timeout applies per-candidate.


## Candidate Sources

A candidate source is a command that outputs JSON - a list of things for Agena to work through. Candidates are evaluated in order and re-generated between runs. Once a candidate has been processed, it won't be retried (tracked via `ignore.log` in your task directory - remove entries to retry them).

Three output formats are supported:

**Strings** - for simple single-value candidates:

```json
["file1.go", "file2.go", "file3.go"]
```

Access in prompt with `$INPUT`.

**Arrays** - when you need multiple values per candidate (e.g. file, line number, message):

```json
[
  ["file.go", "10", "error message"],
  ["other.go", "20", "warning"]
]
```

Access in prompt with `$INPUT[0]`, `$INPUT[1]`, `$INPUT[1:]`.

**Maps** - for self-documenting structured data:

```json
[{ "file": "test.go", "line": 10, "type": "error" }]
```

Access with `$INPUT["file"]`, `$INPUT["line"]`.

## Prompts

Prompts tell Gemini what to do with each candidate. You can either inline them in `task.yaml`:

```yaml
prompt: "Fix this compiler error: $INPUT"
```

Or use a template file for longer prompts:

```yaml
template: "template.txt"
```

**template.txt:**
```
Fix the following compiler error in this codebase.

Error: $INPUT[0]
File: $INPUT[1]
Line: $INPUT[2]

Make the minimal change necessary to resolve the error.
```

### Variable Reference

| Syntax          | Description                          | Example Output             |
| --------------- | ------------------------------------ | -------------------------- |
| `$INPUT`        | Whole input (single items unwrapped) | `"file.go"` or `["a","b"]` |
| `$INPUT[0]`     | Array index (0-based)                | First element              |
| `$INPUT[1]`     | Array index                          | Second element             |
| `$INPUT[1:]`    | Slice from index to end              | `["b","c","d"]`            |
| `$INPUT["key"]` | Map key lookup                       | Value for key              |

## Best-Effort Mode

By default, Agena resets changes if the candidate is still present after Gemini's fix. This makes sense for things like compiler errors where you need exact resolution.

For tasks where partial progress is valuable (refactoring, lint fixes, etc.), set:

```yaml
accept_best_effort: true
```

This commits whatever Gemini produces, regardless of whether the candidate fully resolves.
