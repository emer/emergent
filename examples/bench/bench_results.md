# Benchmark results

## C++ Emergent

Results are total time for 1, 2, 4 threads, on my macbook

```
* SMALL:   2.383   2.248    2.042
* MEDIUM:  2.535   1.895    1.263
* LARGE:  19.627   8.559    8.105
* HUGE:   24.119  11.803   11.897
* GINOR:  35.334  24.768   17.794
```

## Go emergent, per-layer threads, thread pool, optimized range synapse code

Results are total time for 1, 2, 4 threads, on my macbook

```
* SMALL:   1.486   4.297   4.644
* MEDIUM:  2.864   3.312   3.037
* LARGE:  25.09   20.06   16.88
* HUGE:   31.39   23.85   19.53
* GINOR:  42.18   31.29   26.06
```

not too much diff for wt bal off!

## Go emergent, per-layer threads, thread pool

Results are total time for 1, 2, 4 threads, on my macbook

```
* SMALL:  1.2180    4.25328  4.66991
* MEDIUM: 3.392145  3.631261  3.38302
* LARGE:  31.27893  20.91189 17.828935
* HUGE:   42.0238   22.64010  18.838019
* GINOR:  65.67025  35.54374  27.56567
```

## Go emergent, per-layer threads, no thread pool (de-novo threads)

Results are total time for 1, 2, 4 threads, on my macbook

```
* SMALL:  1.2180    3.548349  4.08908
* MEDIUM: 3.392145  3.46302   3.187831
* LARGE:  31.27893  22.20344  18.797924
* HUGE:   42.0238   29.00472  24.53498
* GINOR:  65.67025  45.09239  36.13568
```

# Per Function 

Focusing on the LARGE case:

C++: `emergent -nogui -ni -p leabra_bench.proj epochs=5 pats=20 units=625 n_threads=1`

```
BenchNet_5lay timing report:
function  	time     percent 
Net_Input     8.91    43.1
Net_InInteg	   0.71     3.43
Activation    1.95     9.43
Weight_Change 4.3     20.8
Weight_Updt	   2.85    13.8
Net_InStats	   0.177    0.855
Inhibition    0.00332  0.016
Act_post      1.63     7.87
Cycle_Stats	   0.162    0.781
    total:	   20.7
```

Go: `./bench -epochs 5 -pats 20 -units 625 -threads=1`

```
TimerReport: BenchNet, NThreads: 1
    Function Name  Total Secs    Pct
    ActFmG         2.121      8.223
    AvgMaxAct      0.1003     0.389
    AvgMaxGe       0.1012     0.3922
    DWt            5.069     19.65
    GeFmGeInc      0.3249     1.259
    InhibFmGeAct   0.08498    0.3295
    QuarterFinal   0.003773   0.01463
    SendGeDelta   14.36      55.67
    WtBalFmWt      0.1279     0.4957
    WtFmDWt        3.501     13.58
    Total         25.79
```

```
TimerReport: BenchNet, NThreads: 1
    Function Name    Total Secs    Pct
    ActFmG           2.119     7.074
    AvgMaxAct        0.1        0.3339
    AvgMaxGe        0.102     0.3407
    DWt             5.345     17.84
    GeFmGeInc        0.3348     1.118
    InhibFmGeAct     0.0842     0.2811
    QuarterFinal     0.004    0.01351
    SendGeDelta     17.93     59.87
    WtBalFmWt        0.1701     0.568
    WtFmDWt        3.763     12.56
    Total         29.96
```

* trimmed 4+ sec from SendGeDelta by avoiding range checks using sub-slices
* was very sensitive to size of Synapse struct


