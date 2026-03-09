#!/bin/sh

target=""
for arg in "$@"; do
  case "$arg" in
    -*) ;;
    *) target="$arg" ;;
  esac
done

if [ -z "$target" ]; then
  target="127.0.0.1"
fi

echo "PING $target (127.0.0.1): 56 data bytes"
echo "64 bytes from 127.0.0.1: icmp_seq=0 ttl=64 time=12.3 ms"
sleep 0.05
echo "64 bytes from 127.0.0.1: icmp_seq=1 ttl=64 time=18.4 ms"
sleep 0.05
echo "Request timeout for icmp_seq 2"
sleep 0.05
echo "64 bytes from 127.0.0.1: icmp_seq=3 ttl=64 time=15.1 ms"
sleep 5
