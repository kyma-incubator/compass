#!/usr/bin/env bash

print_warning() {
  echo -e "\033[33m $1 \033[39m"
}

print_error() {
  echo -e "\033[31m $1 \033[39m"
}

print_ok() {
  echo -e "\033[32m $1 \033[39m"
}
