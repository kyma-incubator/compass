#!/usr/bin/env bash
#Script downloaded from https://raw.githubusercontent.com/kyma-project/kyma/1.1.0/installation/scripts/utils.sh

#
# Log is function useful for logs creation.
# It accepts three arguments:
# $1 - text we want to print
# $2 - text color
# $3 - text style
# It will create single log with defined color and font style.
# To specify style without color we have to put 'nc' before style.
# For example:
# log "gophers" magenta bold - will print bold 'gophers' in magenta color.
# log "text" bold - it will print normal text.
# log "text" nc bold - it will print bold text.
# By default log will print normal text like echo command.
# Use source [utils.sh path] to import log function into your script.
#

function log() {
    local exp=$1;
    local color=$2;
    local style=$3;
    local NC='\033[0m'
    if ! [[ ${color} =~ '^[0-9]$' ]] ; then
       case $(echo ${color} | tr '[:upper:]' '[:lower:]') in
        black) color='\e[30m' ;;
        red) color='\e[31m' ;;
        green) color='\e[32m' ;;
        yellow) color='\e[33m' ;;
        blue) color='\e[34m' ;;
        magenta) color='\e[35m' ;;
        cyan) color='\e[36m' ;;
        white) color='\e[37m' ;;
        nc|*) color=${NC} ;; # no color or invalid color
       esac
    fi
    if ! [[ ${style} =~ '^[0-9]$' ]] ; then
        case $(echo ${style} | tr '[:upper:]' '[:lower:]') in
        bold) style='\e[1m' ;;
        underline) style='\e[4m' ;;
        inverted) style='\e[7m' ;;
        *) style="" ;; # no style or invalid style
       esac
    fi
    printf "${color}${style}${exp}${NC}\n"
}

# checkInputParameterValue is a function to check if input parameter is valid
# There HAS to be provided argument:
# $1 - value for input parameter
# for example in installation/cmd/run.sh we can set --vm-driver argument, which has to have a value.

function checkInputParameterValue() {
    if [ -z "${1}" ] || [ "${1:0:2}" == "--" ]; then
        echo "Wrong parameter value"
        echo "Make sure parameter value is neither empty nor start with two hyphens"
        exit 1
    fi
}