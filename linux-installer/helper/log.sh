#!/bin/bash

function logger_start {
    local LOG_FILE="$1"
    echo "" | tee -a "$LOG_FILE"
    exec > >(tee -a ${LOG_FILE} )
    exec 2> >(tee -a ${LOG_FILE} >&2)
}
