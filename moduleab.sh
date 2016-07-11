#!/bin/sh

DIR=`dirname $0`
PIDFILE='moduleab.pid'
EXEC='./moduleab_agent'

cd $DIR

function run() {
  while true
  do
    $EXEC
  done
}

function start(){
  run &
  disown
  echo "ModuleAB Agent Started."
}

function stop(){
  PPID=`cat /proc/$(cat $PIDFILE)/status|grep PPid|awk '{print $2}'`
  if [ '1' != "$PPID" ];then
    kill $PPID
  fi
  kill `cat $PIDFILE`
  echo "ModuleAB Agent Stopped."
}

function restart(){
  stop
  sleep 1
  start
}

if [ 'x' = "${1}x" ]; then
  echo "$0 {start|stop|restart}"
else
  $1
fi
