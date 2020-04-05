#!/bin/sh

basePort=3000
minRacers=2
racers=$1

# check if number of racers given
if [ $# -eq 0 ]
  then
    echo "No arguments supplied"
    exit 1
fi

# racers must atleast be equal to minimum racers
if [ $racers -lt $minRacers ]
then 
    echo "racers atleast ${minRacers}"
    exit 1
fi

# create the command to execute go script parallely
command="/go/bin/line-racer -racers ${racers}"

for i in $(seq 1 $racers)
do
    port=$(( $basePort + $i))
    command="${command} & /go/bin/line-racer -nodeType racer -port ${port}"
done

# execute the command
echo $command
eval "${command}"