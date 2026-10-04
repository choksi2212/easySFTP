# easySFTP benchmark: release v3.8.2

| Field | Value |
|---|---|
| Kind | release (official reference) |
| Version | `v3.8.2` |
| Recorded | 2026-10-04T16:04:15Z |
| Commit | `9f64c5e68d7df8f6cd3463aafc10e9949893a7e1` |
| Workflow run | https://github.com/eiserv/easySFTP/actions/runs/37215368810 |
| Raw data | [release-v3.8.2.json](release-v3.8.2.json) |
| Flat export | [release-v3.8.2.csv](release-v3.8.2.csv) |

## easySFTP benchmark

| Setting | Value |
|---|---|
| Candidate | `v3.8.2 (9f64c5e)` |
| Baseline | `none` |
| Repeats per scenario | 3 |
| Runner | Linux 7.0.0-30-generic, 10 cpu |
| Link profiles | the real line |
| Settings | easySFTP defaults (no advanced.* overrides): connections auto, concurrency auto, request_concurrency auto, retries 2, timeout 30s, mode overlay |

### The link

| Profile | When | RTT p50 | RTT p90 | Handshake | Control 1 stream | Control N streams | Host load |
|---|---|---|---|---|---|---|---|
| baseline | start | 12.86 ms | 13.11 ms | 386.94 ms | 0.41 MiB/s | 1.16 MiB/s | n/a |
| baseline | end | 13.05 ms | 13.51 ms | 406.73 ms | 0.39 MiB/s | 1.1 MiB/s | n/a |

No link shaping was requested: every profile here is the real line.

The control measurement uses `x/crypto/ssh` and `pkg/sftp` directly, never easySFTP's uploader. It separates "the line is slow" from "easySFTP is slow", and a single-stream control close to a scenario's own MiB/s means the run was network bound, where a code delta says nothing.

### Throughput

| Scenario | Build | Profile | Files | Size | Median | Min | Max | MAD | MiB/s | files/s | Retries | Errors | Failed runs | Delta |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| small | candidate | baseline | 300 | 1.2 MiB | 3922 ms | 3582 ms | 3937 ms | 15 ms | 0.3 | 76.49 | 0 | 0 | 0 | - |
| mixed | candidate | baseline | 56 | 11.6 MiB | 11370 ms | 10395 ms | 12185 ms | 815 ms | 1.02 | 4.93 | 0 | 0 | 0 | - |
| large | candidate | baseline | 2 | 32 MiB | 50042 ms | 47644 ms | 52215 ms | 2173 ms | 0.64 | 0.04 | 0 | 0 | 0 | - |

Delta compares each build's median against the `candidate` build **on the same link profile**; negative is faster. MAD is the median absolute deviation of the repeats: a delta smaller than it is inside this host's own noise.

### Resources (median per run)

| Scenario | Build | Profile | User CPU | Sys CPU | CPU % | Peak RSS | Go allocs | GCs | GC pause | Peak goroutines | Net sent |
|---|---|---|---|---|---|---|---|---|---|---|---|
| small | candidate | baseline | 189.64 ms | 241.69 ms | 11.22% | 12.8 MiB | 11.7 MiB | 5 | 1.903644 ms | 106 | 1.4 MiB |
| mixed | candidate | baseline | 206.6 ms | 221.09 ms | 3.62% | 11.8 MiB | 4.4 MiB | 2 | 0.423911 ms | 106 | 11.7 MiB |
| large | candidate | baseline | 436.82 ms | 534.54 ms | 1.98% | 8.8 MiB | 1.2 MiB | 0 | 0 ms | 25 | 32.1 MiB |

### Where the time goes

Phases are wall clock and add up to roughly the run's duration. Operation totals are **cumulative across parallel workers** and are normally larger than the phase they belong to; read them for their share and their per-call cost, never as wall clock.

<details><summary><code>small</code> phases and round-trips</summary>

| Build | Profile | Phase | Wall |
|---|---|---|---|
| candidate | baseline | upload | 3319.03 ms |
| candidate | baseline | connect | 424.76 ms |
| candidate | baseline | cleanup | 55.99 ms |
| candidate | baseline | create_dirs | 55.51 ms |
| candidate | baseline | sweep_stale_temps | 54.78 ms |
| candidate | baseline | local_scan | 4.45 ms |

| Build | Profile | Operation | Count | Cumulative | Avg | p50 | p90 | p99 | Max |
|---|---|---|---|---|---|---|---|---|---|
| candidate | baseline | file_upload | 300 | 187926.91 ms | 626.42 ms | 479.15 ms | 1416.36 ms | 1649.56 ms | 1685.3 ms |
| candidate | baseline | sftp_write | 300 | 56741.66 ms | 189.14 ms | 165.08 ms | 448.64 ms | 596.49 ms | 732.72 ms |
| candidate | baseline | sftp_open | 300 | 27100.33 ms | 90.33 ms | 50.21 ms | 193.59 ms | 466.2 ms | 559.89 ms |
| candidate | baseline | sftp_rename | 300 | 22574.59 ms | 75.25 ms | 44.8 ms | 182.44 ms | 379.85 ms | 557.9 ms |
| candidate | baseline | sftp_chmod | 300 | 20193.29 ms | 67.31 ms | 48.32 ms | 119.69 ms | 399.78 ms | 560.91 ms |
| candidate | baseline | sftp_readdir | 9 | 484.76 ms | 53.86 ms | 53.56 ms | 54.37 ms | 54.37 ms | 54.37 ms |
| candidate | baseline | sftp_mkdirall | 8 | 383.66 ms | 47.96 ms | 48.53 ms | 55.12 ms | 55.12 ms | 55.12 ms |
| candidate | baseline | ssh_connect | 1 | 381.69 ms | 381.69 ms | 381.69 ms | 381.69 ms | 381.69 ms | 381.69 ms |
| candidate | baseline | sftp_realpath | 3 | 39.77 ms | 13.26 ms | 13.25 ms | 13.45 ms | 13.45 ms | 13.45 ms |

</details>

<details><summary><code>mixed</code> phases and round-trips</summary>

| Build | Profile | Phase | Wall |
|---|---|---|---|
| candidate | baseline | upload | 10722.38 ms |
| candidate | baseline | connect | 405.75 ms |
| candidate | baseline | cleanup | 107.79 ms |
| candidate | baseline | sweep_stale_temps | 54.64 ms |
| candidate | baseline | create_dirs | 51.62 ms |
| candidate | baseline | local_scan | 1.22 ms |

| Build | Profile | Operation | Count | Cumulative | Avg | p50 | p90 | p99 | Max |
|---|---|---|---|---|---|---|---|---|---|
| candidate | baseline | file_upload | 56 | 165724.08 ms | 2959.36 ms | 2734.07 ms | 5279.37 ms | 10721.72 ms | 10721.72 ms |
| candidate | baseline | sftp_write | 56 | 75882.98 ms | 1355.05 ms | 399.4 ms | 3276.13 ms | 8987.95 ms | 8987.95 ms |
| candidate | baseline | sftp_rename | 56 | 7581.77 ms | 135.39 ms | 97.2 ms | 276.47 ms | 651.88 ms | 651.88 ms |
| candidate | baseline | sftp_chmod | 56 | 7144.51 ms | 127.58 ms | 88.75 ms | 294.75 ms | 429.99 ms | 429.99 ms |
| candidate | baseline | sftp_open | 56 | 1268.26 ms | 22.65 ms | 21.48 ms | 32.19 ms | 39.84 ms | 39.84 ms |
| candidate | baseline | sftp_readdir | 9 | 484.85 ms | 53.87 ms | 54 ms | 54.33 ms | 54.33 ms | 54.33 ms |
| candidate | baseline | ssh_connect | 1 | 367.14 ms | 367.14 ms | 367.14 ms | 367.14 ms | 367.14 ms | 367.14 ms |
| candidate | baseline | sftp_mkdirall | 8 | 366.84 ms | 45.86 ms | 46.65 ms | 50.89 ms | 50.89 ms | 50.89 ms |
| candidate | baseline | sftp_realpath | 3 | 39.01 ms | 13 ms | 12.86 ms | 13.08 ms | 13.08 ms | 13.08 ms |

</details>

<details><summary><code>large</code> phases and round-trips</summary>

| Build | Profile | Phase | Wall |
|---|---|---|---|
| candidate | baseline | upload | 49479.32 ms |
| candidate | baseline | connect | 418.95 ms |
| candidate | baseline | sweep_stale_temps | 53.28 ms |
| candidate | baseline | create_dirs | 42.64 ms |
| candidate | baseline | cleanup | 27.03 ms |
| candidate | baseline | local_scan | 0.25 ms |

| Build | Profile | Operation | Count | Cumulative | Avg | p50 | p90 | p99 | Max |
|---|---|---|---|---|---|---|---|---|---|
| candidate | baseline | file_upload | 2 | 96616.46 ms | 48308.23 ms | 49479.16 ms | 49479.16 ms | 49479.16 ms | 49479.16 ms |
| candidate | baseline | sftp_write | 2 | 96117.75 ms | 48058.88 ms | 49025.96 ms | 49025.96 ms | 49025.96 ms | 49025.96 ms |
| candidate | baseline | ssh_connect | 1 | 378.83 ms | 378.83 ms | 378.83 ms | 378.83 ms | 378.83 ms | 378.83 ms |
| candidate | baseline | sftp_readdir | 3 | 159.34 ms | 53.11 ms | 53.1 ms | 53.18 ms | 53.18 ms | 53.18 ms |
| candidate | baseline | sftp_mkdirall | 2 | 83.09 ms | 41.55 ms | 42.52 ms | 42.52 ms | 42.52 ms | 42.52 ms |
| candidate | baseline | sftp_realpath | 3 | 39.45 ms | 13.15 ms | 13.11 ms | 13.65 ms | 13.65 ms | 13.65 ms |
| candidate | baseline | sftp_open | 2 | 30.6 ms | 15.3 ms | 16.38 ms | 16.38 ms | 16.38 ms | 16.38 ms |
| candidate | baseline | sftp_rename | 2 | 29.97 ms | 14.99 ms | 15.22 ms | 15.22 ms | 15.22 ms | 15.22 ms |
| candidate | baseline | sftp_chmod | 2 | 26.68 ms | 13.34 ms | 13.61 ms | 13.61 ms | 13.61 ms | 13.61 ms |

</details>

### Delete sweeps

The pre-clean before every measured run wipes the tree the previous repeat left behind, which makes it a pure delete sweep. It costs no extra time (it has always run) and its numbers never enter the upload tables above. Sweeps that found an empty directory are not counted.

| Scenario | Build | Profile | Sweeps | Files deleted | Median | files/s | remote_scan | delete_sweep |
|---|---|---|---|---|---|---|---|---|
| large | candidate | baseline | 2 | 2 | 559 ms | 3.58 | 104.05 ms | 34.31 ms |
| mixed | candidate | baseline | 2 | 56 | 674 ms | 83.09 | 107.56 ms | 133.75 ms |
| small | candidate | baseline | 2 | 300 | 1288 ms | 232.92 | 125.47 ms | 692.59 ms |

| Scenario | Build | Profile | Operation | Count | Cumulative | p50 | p90 | p99 | Max |
|---|---|---|---|---|---|---|---|---|---|
| large | candidate | baseline | ssh_connect | 1 | 355.19 ms | 355.19 ms | 355.19 ms | 355.19 ms | 355.19 ms |
| large | candidate | baseline | sftp_readdir | 4 | 207.66 ms | 51.92 ms | 51.94 ms | 51.94 ms | 51.94 ms |
| large | candidate | baseline | sftp_rmdir | 2 | 33.24 ms | 18.27 ms | 18.27 ms | 18.27 ms | 18.27 ms |
| large | candidate | baseline | sftp_remove | 2 | 30.36 ms | 15.86 ms | 15.86 ms | 15.86 ms | 15.86 ms |
| mixed | candidate | baseline | sftp_remove | 56 | 3122.79 ms | 56.14 ms | 90.63 ms | 97.25 ms | 97.25 ms |
| mixed | candidate | baseline | sftp_readdir | 10 | 535.59 ms | 53.92 ms | 54.83 ms | 54.83 ms | 54.83 ms |
| mixed | candidate | baseline | ssh_connect | 1 | 364.51 ms | 364.51 ms | 364.51 ms | 364.51 ms | 364.51 ms |
| mixed | candidate | baseline | sftp_rmdir | 8 | 195.39 ms | 24.95 ms | 33.07 ms | 33.07 ms | 33.07 ms |
| small | candidate | baseline | sftp_remove | 300 | 37380.34 ms | 121.83 ms | 173.07 ms | 181.6 ms | 183.88 ms |
| small | candidate | baseline | sftp_readdir | 10 | 610.44 ms | 58.08 ms | 70.29 ms | 70.29 ms | 70.29 ms |
| small | candidate | baseline | ssh_connect | 1 | 395.63 ms | 395.63 ms | 395.63 ms | 395.63 ms | 395.63 ms |
| small | candidate | baseline | sftp_rmdir | 8 | 193.45 ms | 24.61 ms | 31.88 ms | 31.88 ms | 31.88 ms |

Data only: these numbers set no threshold and fail no build. Collected to evaluate the single-connection ceiling discussed in issue #158 and to show where a run spends its time.
