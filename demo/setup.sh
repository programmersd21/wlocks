#!/usr/bin/env bash
DIR=/tmp/wlocks_demo
rm -rf "$DIR" 2>/dev/null
mkdir -p "$DIR"
echo "locked by process A (read)" > "$DIR/target.lock"
echo "locked by process B (write)" >> "$DIR/target.lock"
echo "locked by process C (multi-fd)" >> "$DIR/target.lock"
echo "other file data" > "$DIR/other.lock"
nohup tail -f "$DIR/target.lock" > /dev/null 2>&1 &
nohup bash -c 'exec 5>'"$DIR/target.lock"'; while true; do sleep 10; done' > /dev/null 2>&1 &
nohup bash -c 'exec 5<'"$DIR/target.lock"'; exec 6<'"$DIR/other.lock"'; while true; do sleep 10; done' > /dev/null 2>&1 &
echo "ready"
