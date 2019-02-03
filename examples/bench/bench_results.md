# Benchmark results

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

## C++ Emergent

Results are total time for 1, 2, 4 threads, on my macbook

```
* SMALL:  2.38307  2.2484   2.04244
* MEDIUM: 2.53459  1.8954   1.2634
* LARGE:  19.6275  8.55913  8.10503
* HUGE:   24.1191  11.8032  11.8969
* GINOR:  35.3342  24.768   17.7942
```


