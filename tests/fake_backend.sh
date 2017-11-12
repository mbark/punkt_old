#!/usr/bin/env bash

if [ "$1" == "list" ] ; then
    echo "package1"
    echo "package2"
    echo "package3"
    echo "package4"
    exit 0
fi

if [ "$1" == "update" ] ; then
    echo "updated"
    exit 0
fi

if [ "$1" == "install" ] ; then
    echo "installed"
    echo 0
fi
