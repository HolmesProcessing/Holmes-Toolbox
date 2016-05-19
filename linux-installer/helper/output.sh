#!/bin/bash

BG_RED='\033[7;49;91m'
RED='\033[0;49;91m'
GREEN='\033[0;49;92m'
MAGENTA='\033[0;49;95m'
CYAN='\033[0;49;96m'
ENDC='\033[0m'

function error_bg {
    >&2 printf "${BG_RED}%s${ENDC}\n" "$1"
}
function error {
    >&2 printf "${RED}%s${ENDC}\n" "$1"
}
function info {
    printf "${CYAN}%s${ENDC}\n" "$1"
}
function tolower {
    x=$(echo "$1" | tr '[:upper:]' '[:lower:]')
    echo $x
}
function readinput {
    read -e -p "${MAGENTA}$1: ${ENDC}" INPUT
    INPUT=$(tolower $INPUT)
    echo "$INPUT"
}
