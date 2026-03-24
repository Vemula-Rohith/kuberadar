# KubeRadar

KubeRadar is a Kubernetes debugging and diagnostics CLI that helps developers diagnose and debug issues in clusters.

## Installation

```bash
go install github.com/Vemula-Rohith/kuberadar/cmd/kuberadar@latest
```

Or build from source:

```bash
git clone https://github.com/Vemula-Rohith/kuberadar
cd kuberadar
go build -o kuberadar ./cmd/kuberadar
```

## Usage

### Commands

- **`kuberadar sweep`** — Sweep all pods in a namespace for issues
- **`kuberadar pod [name]`** — Diagnose a specific pod or all pods in namespace (use `--diagnose` for full evidence and recommendation)
- **`kuberadar deployment [name]`** — Diagnose a deployment by checking its pods
- **`kuberadar explain [issue-id]`** — Explain an issue ID and how to resolve it (self-documenting)
- **`kuberadar version`** — Print the version

### Flags

- `--kubeconfig` — Path to kubeconfig file
- `-n, --namespace` — Kubernetes namespace (default from kubeconfig context)
- `-o, --output` — Output format: `table` (default), `json`

### Examples

```bash
# Sweep all pods in default namespace
kuberadar sweep

# Sweep pods in a specific namespace
kuberadar sweep -n my-namespace

# Diagnose a specific pod (summary)
kuberadar pod my-pod -n my-namespace

# Full diagnosis with evidence and recommendation
kuberadar pod my-pod -n my-namespace --diagnose

# Output as JSON for CI
kuberadar sweep -o json

# Explain an issue ID
kuberadar explain KR003
```

## Exit codes

By default, **exit 0 means the command completed successfully** (cluster was reached and output was produced). Finding issues is normal output, not a failed command.

Use **`--fail-on-issues`** for CI-style gates:

| Code | Meaning (only with `--fail-on-issues`) |
|------|----------------------------------------|
| 0    | No issues                              |
| 1    | Warnings only                          |
| 2    | Critical issues                        |

Example: `kuberadar sweep --fail-on-issues || exit 1` fails the pipeline when any issue is reported.

## Issue IDs

| ID   | Description           | Severity |
|------|-----------------------|----------|
| KR001| OOMKilled             | Critical |
| KR002| CrashLoopBackOff      | Critical |
| KR003| ImagePullBackOff      | Critical |
| KR004| UnschedulablePod      | Warning  |

## Architecture

- **CLI** — Cobra commands and flags
- **Engine** — Orchestrates diagnostics via `Run(scope)`
- **Detectors** — Problem detection logic (crashloop, OOM, image pull, scheduling)
- **Providers** — Thin client-go wrappers for Kubernetes API
- **Context Builder** — Builds context per resource (events, node)
