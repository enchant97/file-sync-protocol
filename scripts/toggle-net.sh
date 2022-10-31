#!/bin/bash
#
# Toggle network interface state with a delay
#

if [ "$EUID" -ne 0 ]
then
    echo "this must be run as root"
    exit 1
elif [ "$#" -ne 3 ]
then
    echo "missing required arguments"
    echo "usage: $0 <interface> <up_time> <down_time>"
    exit 1
fi


interface=$1
up_time=$2
down_time=$3


log_datetime() {
    echo $(date --rfc-3339 ns)
}


echo "$(log_datetime)::setting ${interface} up"
ip link set ${interface} up

echo "$(log_datetime)::waiting for ${up_time} seconds"
sleep ${up_time}

echo "$(log_datetime)::setting ${interface} down"
ip link set ${interface} down

echo "$(log_datetime)::waiting for ${down_time} seconds"
sleep ${down_time}

echo "$(log_datetime)::setting ${interface} up"
ip link set ${interface} up
