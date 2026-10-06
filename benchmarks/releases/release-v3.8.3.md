# easySFTP benchmark: release v3.8.3

| Field | Value |
|---|---|
| Kind | release (official reference) |
| Version | `v3.8.3` |
| Recorded | 2026-10-06T11:50:05Z |
| Commit | `ca27939f1bbe5e2c1cddb35d8827d7be1e09f7d0` |
| Workflow run | https://github.com/eiserv/easySFTP/actions/runs/37458989533 |
| Raw data | [release-v3.8.3.json](release-v3.8.3.json) |
| Flat export | [release-v3.8.3.csv](release-v3.8.3.csv) |

## easySFTP benchmark

| Setting | Value |
|---|---|
| Candidate | `v3.8.3 (ca27939)` |
| Baseline | `none` |
| Repeats per scenario | 3 |
| Runner | Linux 7.0.0-30-generic, 10 cpu |
| Link profiles | the real line |
| Settings | easySFTP defaults (no advanced.* overrides): connections auto, concurrency auto, request_concurrency auto, retries 2, timeout 30s, mode overlay |

### The link

| Profile | When | RTT p50 | RTT p90 | Handshake | Control 1 stream | Control N streams | Host load |
|---|---|---|---|---|---|---|---|
| baseline | start | 12.91 ms | 13.05 ms | 412.78 ms | 0.42 MiB/s | 1.2 MiB/s | n/a |
| baseline | end | 12.76 ms | 13.54 ms | 371.43 ms | 0.42 MiB/s | 1.21 MiB/s | n/a |

No link shaping was requested: every profile here is the real line.

The control measurement uses `x/crypto/ssh` and `pkg/sftp` directly, never easySFTP's uploader. It separates "the line is slow" from "easySFTP is slow", and a single-stream control close to a scenario's own MiB/s means the run was network bound, where a code delta says nothing.

### Throughput

| Scenario | Build | Profile | Files | Size | Median | Min | Max | MAD | MiB/s | files/s | Retries | Errors | Failed runs | Delta |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| small | candidate | baseline | 300 | 1.2 MiB | 3649 ms | 3229 ms | 3668 ms | 19 ms | 0.32 | 82.21 | 0 | 0 | 0 | - |
| mixed | candidate | baseline | 56 | 11.6 MiB | 10607 ms | 10238 ms | 11172 ms | 369 ms | 1.1 | 5.28 | 0 | 0 | 0 | - |
| large | candidate | baseline | 2 | 32 MiB | 45747 ms | 44707 ms | 48114 ms | 1040 ms | 0.7 | 0.04 | 0 | 0 | 0 | - |

Delta compares each build's median against the `candidate` build **on the same link profile**; negative is faster. MAD is the median absolute deviation of the repeats: a delta smaller than it is inside this host's own noise.

### Resources (median per run)

| Scenario | Build | Profile | User CPU | Sys CPU | CPU % | Peak RSS | Go allocs | GCs | GC pause | Peak goroutines | Net sent |
|---|---|---|---|---|---|---|---|---|---|---|---|
| small | candidate | baseline | 198.26 ms | 238.51 ms | 13.25% | 13.8 MiB | 12.4 MiB | 6 | 1.271631 ms | 244 | 1.4 MiB |
| mixed | candidate | baseline | 229.41 ms | 236.26 ms | 4.55% | 14.2 MiB | 5.4 MiB | 3 | 1.322031 ms | 1091 | 11.7 MiB |
| large | candidate | baseline | 470.9 ms | 490.05 ms | 2.11% | 8.8 MiB | 1.3 MiB | 0 | 0 ms | 159 | 32.1 MiB |

### Where the time goes

Phases are wall clock and add up to roughly the run's duration. Operation totals are **cumulative across parallel workers** and are normally larger than the phase they belong to; read them for their share and their per-call cost, never as wall clock.

<details><summary><code>small</code> phases and round-trips</summary>

| Build | Profile | Phase | Wall |
|---|---|---|---|
| candidate | baseline | upload | 3049.01 ms |
| candidate | baseline | connect | 385.35 ms |
| candidate | baseline | cleanup | 66.19 ms |
| candidate | baseline | sweep_stale_temps | 55.03 ms |
| candidate | baseline | create_dirs | 54.49 ms |
| candidate | baseline | local_scan | 3.8 ms |

| Build | Profile | Operation | Count | Cumulative | Avg | p50 | p90 | p99 | Max |
|---|---|---|---|---|---|---|---|---|---|
| candidate | baseline | file_upload | 300 | 176798.89 ms | 589.33 ms | 343.28 ms | 1467.92 ms | 2226.24 ms | 2231.22 ms |
| candidate | baseline | sftp_write | 300 | 48570.61 ms | 161.9 ms | 95.69 ms | 341.28 ms | 402.03 ms | 467.31 ms |
| candidate | baseline | sftp_open | 300 | 21274.8 ms | 70.92 ms | 47.64 ms | 174.2 ms | 325.81 ms | 381.02 ms |
| candidate | baseline | sftp_chmod | 300 | 17053.47 ms | 56.84 ms | 49.19 ms | 79.25 ms | 316.37 ms | 367.15 ms |
| candidate | baseline | sftp_rename | 300 | 15049.82 ms | 50.17 ms | 36.97 ms | 95.24 ms | 265.65 ms | 310.01 ms |
| candidate | baseline | sftp_readdir | 9 | 483.52 ms | 53.72 ms | 54.49 ms | 54.73 ms | 54.73 ms | 54.73 ms |
| candidate | baseline | sftp_mkdirall | 8 | 385.96 ms | 48.25 ms | 48.5 ms | 54.24 ms | 54.24 ms | 54.24 ms |
| candidate | baseline | ssh_connect | 1 | 346.79 ms | 346.79 ms | 346.79 ms | 346.79 ms | 346.79 ms | 346.79 ms |
| candidate | baseline | sftp_realpath | 3 | 38.52 ms | 12.84 ms | 12.78 ms | 13 ms | 13 ms | 13 ms |

</details>

<details><summary><code>mixed</code> phases and round-trips</summary>

| Build | Profile | Phase | Wall |
|---|---|---|---|
| candidate | baseline | upload | 9990.5 ms |
| candidate | baseline | connect | 402.55 ms |
| candidate | baseline | cleanup | 106.94 ms |
| candidate | baseline | sweep_stale_temps | 66.04 ms |
| candidate | baseline | create_dirs | 52.61 ms |
| candidate | baseline | local_scan | 1.41 ms |

| Build | Profile | Operation | Count | Cumulative | Avg | p50 | p90 | p99 | Max |
|---|---|---|---|---|---|---|---|---|---|
| candidate | baseline | file_upload | 56 | 216384.07 ms | 3864 ms | 3863.05 ms | 6041.25 ms | 9989.47 ms | 9989.47 ms |
| candidate | baseline | sftp_write | 56 | 94599.1 ms | 1689.27 ms | 1223.53 ms | 2855.66 ms | 8254.04 ms | 8254.04 ms |
| candidate | baseline | sftp_chmod | 56 | 23857.89 ms | 426.03 ms | 472.55 ms | 906.83 ms | 1224.93 ms | 1224.93 ms |
| candidate | baseline | sftp_rename | 56 | 19889.01 ms | 355.16 ms | 401.93 ms | 999.26 ms | 1052.65 ms | 1052.65 ms |
| candidate | baseline | sftp_open | 56 | 1190.38 ms | 21.26 ms | 20.99 ms | 27.48 ms | 34.52 ms | 34.52 ms |
| candidate | baseline | sftp_readdir | 9 | 592.5 ms | 65.83 ms | 64.58 ms | 65.91 ms | 65.91 ms | 65.91 ms |
| candidate | baseline | sftp_mkdirall | 8 | 374.95 ms | 46.87 ms | 47.22 ms | 52.36 ms | 52.36 ms | 52.36 ms |
| candidate | baseline | ssh_connect | 1 | 364.01 ms | 364.01 ms | 364.01 ms | 364.01 ms | 364.01 ms | 364.01 ms |
| candidate | baseline | sftp_realpath | 3 | 39.21 ms | 13.07 ms | 13.05 ms | 13.2 ms | 13.2 ms | 13.2 ms |

</details>

<details><summary><code>large</code> phases and round-trips</summary>

| Build | Profile | Phase | Wall |
|---|---|---|---|
| candidate | baseline | upload | 45201.88 ms |
| candidate | baseline | connect | 398.51 ms |
| candidate | baseline | sweep_stale_temps | 52.4 ms |
| candidate | baseline | create_dirs | 44.37 ms |
| candidate | baseline | cleanup | 26.91 ms |
| candidate | baseline | local_scan | 0.42 ms |

| Build | Profile | Operation | Count | Cumulative | Avg | p50 | p90 | p99 | Max |
|---|---|---|---|---|---|---|---|---|---|
| candidate | baseline | file_upload | 2 | 84268.72 ms | 42134.36 ms | 45200.76 ms | 45200.76 ms | 45200.76 ms | 45200.76 ms |
| candidate | baseline | sftp_write | 2 | 83807.74 ms | 41903.87 ms | 45157.47 ms | 45157.47 ms | 45157.47 ms | 45157.47 ms |
| candidate | baseline | ssh_connect | 1 | 359.75 ms | 359.75 ms | 359.75 ms | 359.75 ms | 359.75 ms | 359.75 ms |
| candidate | baseline | sftp_readdir | 3 | 156.09 ms | 52.03 ms | 51.92 ms | 52.3 ms | 52.3 ms | 52.3 ms |
| candidate | baseline | sftp_mkdirall | 2 | 85.96 ms | 42.98 ms | 44.25 ms | 44.25 ms | 44.25 ms | 44.25 ms |
| candidate | baseline | sftp_realpath | 3 | 38.74 ms | 12.91 ms | 12.86 ms | 13.03 ms | 13.03 ms | 13.03 ms |
| candidate | baseline | sftp_open | 2 | 30.47 ms | 15.24 ms | 16.04 ms | 16.04 ms | 16.04 ms | 16.04 ms |
| candidate | baseline | sftp_rename | 2 | 30.47 ms | 15.24 ms | 15.55 ms | 15.55 ms | 15.55 ms | 15.55 ms |
| candidate | baseline | sftp_chmod | 2 | 26.35 ms | 13.18 ms | 13.37 ms | 13.37 ms | 13.37 ms | 13.37 ms |

</details>

### Delete sweeps

The pre-clean before every measured run wipes the tree the previous repeat left behind, which makes it a pure delete sweep. It costs no extra time (it has always run) and its numbers never enter the upload tables above. Sweeps that found an empty directory are not counted.

| Scenario | Build | Profile | Sweeps | Files deleted | Median | files/s | remote_scan | delete_sweep |
|---|---|---|---|---|---|---|---|---|
| large | candidate | baseline | 2 | 2 | 555 ms | 3.6 | 105.45 ms | 35.24 ms |
| mixed | candidate | baseline | 2 | 56 | 771 ms | 72.63 | 109.66 ms | 224.89 ms |
| small | candidate | baseline | 2 | 300 | 1502 ms | 199.73 | 123.58 ms | 852.67 ms |

| Scenario | Build | Profile | Operation | Count | Cumulative | p50 | p90 | p99 | Max |
|---|---|---|---|---|---|---|---|---|---|
| large | candidate | baseline | ssh_connect | 1 | 346.2 ms | 346.2 ms | 346.2 ms | 346.2 ms | 346.2 ms |
| large | candidate | baseline | sftp_readdir | 4 | 210.38 ms | 52.99 ms | 53.12 ms | 53.12 ms | 53.12 ms |
| large | candidate | baseline | sftp_remove | 2 | 33.94 ms | 17.41 ms | 17.41 ms | 17.41 ms | 17.41 ms |
| large | candidate | baseline | sftp_rmdir | 2 | 33.43 ms | 17.54 ms | 17.54 ms | 17.54 ms | 17.54 ms |
| mixed | candidate | baseline | sftp_remove | 56 | 5630.35 ms | 104.12 ms | 170.56 ms | 181.45 ms | 181.45 ms |
| mixed | candidate | baseline | sftp_readdir | 10 | 537.58 ms | 54.13 ms | 56.6 ms | 56.6 ms | 56.6 ms |
| mixed | candidate | baseline | ssh_connect | 1 | 365.38 ms | 365.38 ms | 365.38 ms | 365.38 ms | 365.38 ms |
| mixed | candidate | baseline | sftp_rmdir | 8 | 216.57 ms | 29.12 ms | 38.75 ms | 38.75 ms | 38.75 ms |
| small | candidate | baseline | sftp_remove | 300 | 42726.08 ms | 150.96 ms | 187.88 ms | 236.03 ms | 238.03 ms |
| small | candidate | baseline | sftp_readdir | 10 | 603.5 ms | 57.23 ms | 69.92 ms | 69.92 ms | 69.92 ms |
| small | candidate | baseline | ssh_connect | 1 | 372.1 ms | 372.1 ms | 372.1 ms | 372.1 ms | 372.1 ms |
| small | candidate | baseline | sftp_rmdir | 8 | 192.38 ms | 25.65 ms | 31.43 ms | 31.43 ms | 31.43 ms |

Data only: these numbers set no threshold and fail no build. Collected to evaluate the single-connection ceiling discussed in issue #158 and to show where a run spends its time.
