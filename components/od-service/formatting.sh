# definitions for color logging in Open Discovery Service scripts

export CYAN='\033[0;36m'
export GREEN='\033[0;32m'
export RED='\033[0;31m'
export NC='\033[0m'

log_section() {
  echo -e "${GREEN}------------------------------------------------------------------------------------------${NC}"
  echo -e "${GREEN}${1}${NC}"
  echo -e "${GREEN}------------------------------------------------------------------------------------------${NC}"
}

log_info() {
  echo -e "${CYAN}${1}${NC}";
}

log_error() {
  # print on stderr
  >&2 echo -e "${RED}$(date +"%Y-%m-%d %H:%M") ERROR: $1${NC}";
}