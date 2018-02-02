#!/bin/bash

for i in `seq 0 9`; do cat case1-mnist-OFF-20-1-ON-400-round_$i/mnist-case1-pass0.log|awk -F, '{if ($1<=550) {a=a+$2; b=b+1}} END {print a/b}'; done > /tmp/cpu_off.txt
cat /tmp/cpu_off.txt|awk '{a+=$1} END{print a/10}' >> /tmp/cpu_off.txt
for i in `seq 0 9`; do cat case1-mnist-ON-20-1-ON-400-round_$i/mnist-case1-pass0.log|awk -F, '{if ($1<=550) {a=a+$2; b=b+1}} END {print a/b}'; done > /tmp/cpu_on.txt
cat /tmp/cpu_on.txt|awk '{a+=$1} END{print a/10}' >> /tmp/cpu_on.txt
cat case1-mnist-OFF-20-1-ON-400-round_*/mnist-case1-result.csv|grep -v PASS|grep -v AVG|awk -F'|' 'BEGIN {a=0} {print a"|"$3"|"; a+=1; b+=$3} END{print "AVG|"b/(a+1)"|"}' > /tmp/pending_off.txt
cat case1-mnist-ON-20-1-ON-400-round_*/mnist-case1-result.csv|grep -v PASS|grep -v AVG|awk -F'|' 'BEGIN {a=0} {print a"|"$3"|"; a+=1; b+=$3} END{print "AVG|"b/(a+1)}"|"}' > /tmp/pending_on.txt

echo "# case 1 autoscaling on"
echo 'PASS|AVG PENDING TIME|CLUSTER CPU UTILS'
echo '---|---|---'
paste /tmp/pending_on.txt /tmp/cpu_on.txt

echo
echo "# case 1 autoscaling off"
echo 'PASS|AVG PENDING TIME|CLUSTER CPU UTILS'
echo '---|---|---'
paste /tmp/pending_off.txt /tmp/cpu_off.txt

for i in `seq 0 9`; do cat case2-mnist-OFF-6-1-ON-400-round_$i/mnist-case2.log|awk -F, '{if ($1<=550) {a=a+$2; b=b+1}} END {print a/b}'; done > /tmp/cpu_off.txt
cat /tmp/cpu_off.txt|awk '{a+=$1} END{print a/10}' >> /tmp/cpu_off.txt
for i in `seq 0 9`; do cat case2-mnist-ON-6-1-ON-400-round_$i/mnist-case2.log|awk -F, '{if ($1<=550) {a=a+$2; b=b+1}} END {print a/b}'; done > /tmp/cpu_on.txt
cat /tmp/cpu_on.txt|awk '{a+=$1} END{print a/10}' >> /tmp/cpu_on.txt
cat case2-mnist-OFF-6-1-ON-400-round_*/mnist-case1-result.csv|grep -v PASS|grep -v AVG|awk -F'|' 'BEGIN {a=0} {print a"|"$3"|"; a+=1; b+=$3} END{print "AVG|"b/(a+1)"|"}' > /tmp/pending_off.txt
cat case2-mnist-ON-6-1-ON-400-round_*/mnist-case1-result.csv|grep -v PASS|grep -v AVG|awk -F'|' 'BEGIN {a=0} {print a"|"$3"|"; a+=1; b+=$3} END{print "AVG|"b/(a+1)"|"}' > /tmp/pending_on.txt

echo
echo "# case 2 autoscaling on"
echo 'PASS|AVG PENDING TIME|CLUSTER CPU UTILS'
echo '---|---|---'
paste /tmp/pending_on.txt /tmp/cpu_on.txt

echo
echo "# case 2 autoscaling off"
echo 'PASS|AVG PENDING TIME|CLUSTER CPU UTILS'
echo '---|---|---'
paste /tmp/pending_off.txt /tmp/cpu_off.txt

for i in `seq 0 9`; do cat case2-mnist-OFF-6-1-ON-400-round_*/mnist-case2.log |awk -F, '{if ($1>=300 && $1<=370) {a=a+$2; b=b+1}} END {print a/b}'; done > /tmp/off-peak-util-off.txt
for i in `seq 0 9`; do cat case2-mnist-ON-6-1-ON-400-round_*/mnist-case2.log |awk -F, '{if ($1>=300 && $1<=370) {a=a+$2; b=b+1}} END {print a/b}'; done > /tmp/off-peak-util-on.txt
echo
echo "# case 2 autoscaling on, average util during off-peak time"
echo "Off-peak (300s - 370s) average cluster utilization:"
cat /tmp/off-peak-util-on.txt | awk '{a+=$1; b+=1} END {print a/b}'

echo
echo "# case 2 autoscaling off, average util during off-peak time"
echo "Off-peak (300s - 370s) average cluster utilization:"
cat /tmp/off-peak-util-off.txt | awk '{a+=$1; b+=1} END {print a/b}'
