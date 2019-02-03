#!/bin/bash

exe=./bench

echo " "
echo "=============================================================="
echo "SMALL Network (5 x 25 units)"
$exe -epochs 10 -pats 100 -units 25 $*
echo " "
echo "=============================================================="
echo "MEDIUM Network (5 x 100 units)"
$exe -epochs 3 -pats 100 -units 100 $*
echo " "
echo "=============================================================="
echo "LARGE Network (5 x 625 units)"
$exe -epochs 5 -pats 20 -units 625 $*
echo " "
echo "=============================================================="
echo "HUGE Network (5 x 1024 units)"
$exe -epochs 5 -pats 10 -units 1024 $*
echo " "
echo "=============================================================="
echo "GINORMOUS Network (5 x 2048 units)"
$exe -epochs 2 -pats 10 -units 2048 $*
# echo " "
# echo "=============================================================="
# echo "GAZILIOUS Network (5 x 4096 units)"
# $exe -nogui -ni -p leabra_bench.proj epochs=1 pats=10 units=4096 $*

