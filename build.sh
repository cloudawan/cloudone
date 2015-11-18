contains() {
  string="$1"
  substring="$2"
  if test "${string#*$substring}" != "$string"
  then
    return 0
  else
    return 1
  fi
}

is_go_existing() {
  go_version_response=$(go version)
  contains $go_version_response "go version"
  if [ $? -eq 1 ]; then
    return 1
  else
    return 0
  fi
}

install_go() {
  # Install Go 1.5.1
  wget https://storage.googleapis.com/golang/go1.5.1.linux-amd64.tar.gz
  tar -C /usr/local -xzf go1.5.1.linux-amd64.tar.gz
  
  # Go bin path
  export PATH=$PATH:/usr/local/go/bin
}

set_temporarily_go_path() {
  previous_go_path=$GOPATH
  mkdir -p /tmp/go
  # Export Go path
  export GOPATH=/tmp/go
}

unset_temporarily_go_path() {
  GOPATH=$previous_go_path
  # Clean temporarily go path
  rm -rf /tmp/go
}

build_and_package() {
  # Get self and dependent packagess
  go get github.com/cloudawan/cloudone

  # Build
  go build
  mv cloudone docker/cloudone/
  find ! -wholename './docker/*' ! -wholename './docker' ! -wholename '.' -exec rm -rf {} +
  mv docker/version version
  mv docker/environment environment
}

is_go_existing
if [ $? -eq 0 ]; then
  install_go
fi

set_temporarily_go_path
build_and_package
unset_temporarily_go_path
