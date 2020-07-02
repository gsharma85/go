#!/bin/sh
if [ $# -ne 2 ]; then 
	echo "Usage: ssh-add-pass keyfile passfile"
	exit 1
fi

eval $( ssh-agent -s )

pass=$( cat $2 )

expect << EOF
  spawn ssh-add $1
  expect "Enter passphrase"
  send "$pass\r"
  expect eof
EOF

go get -u github.com/gsharma85/go/sensor/cmd/sensor

