# vm-patch-metrics

vm-patch-metrics is a helper tool for VictoriaMetrics.
I wrote it for removing "bad" points from metrics without losing the rest of the metrics.

What it does:
1. Exports the matching metrics to a .jsonl file
2. Removes datapoints from each line that fall within the `-remove-start` and `-remove-end` times, writing them
   to a new .jsonl file
3. **Deletes** the matching metrics from VictoriaMetrics. (For all time!)
4. Imports the metrics without the removed dates back into VictoriaMetrics.

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

Example of removing December 1-4, 2022 data from data exported since Jan 1, 2020:

```./vm-replace-metrics -url http://localhost:8428
-user user -password password
-match 'my_cool_metric{mylabel="a", otherlabel=~"b_.*"}'
-export-start 2020-01-01T00:00:00+05:00
-remove-start 2022-12-01T00:00:00Z05:00
-remove-end 2022-12-05T00:00:00Z05:00
```