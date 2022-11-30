# vm-patch-metrics

vm-patch-metrics is a helper tool for VictoriaMetrics.
I wrote it for removing "bad" points from metrics without losing the rest of the metrics.

## Usage:

    Usage of vm-replace-metrics:
        -export-end string
            End time for the exported metrics (default: current time)
        -export-start string
            Start time for the exported metrics
        -file string
            File path to export metrics to (default "./metrics.jsonl")
        -match string
            Metric expression to export from VM
        -password string
            VM user password to authenticate
        -remove-end string
            End time of the points to remove from exported metrics (default: current time)
        -remove-start string
            Start time of the points to remove from exported metrics
        -url string
            VM url (default "http://localhost:8428")
        -user string
            VM user to authenticate