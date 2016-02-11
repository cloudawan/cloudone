#!/bin/bash

# Check whether Etcd is up
# Parameter
retry_amount=120
etcd_endpoint_list_with_quotation=${ETCD_ENDPOINT//,}
etcd_endpoint_list=${etcd_endpoint_list_with_quotation//\"}

is_etcd_up() {
  for etcd_endpoint in $etcd_endpoint_list
  do
    etcd_response=$(curl -m 1 "$etcd_endpoint")

    if [[ $etcd_response == *"404 page not found"* ]]; then
      return 1
    fi
  done
  return 0
}

# Etcd
for ((i=0;i<$retry_amount;i++))
do
  echo "ping $i times to Etcd"
  is_etcd_up
  etcd_result=$?
  if [ $etcd_result == 1 ]; then
	break
  fi
  sleep 1
done

if [ $i == $retry_amount ]; then
  echo "Could not get ping response from Etcd"
  exit -1
fi






# Use environment
# Use different delimiter since URL contains slash
sed -i "s#{{ETCD_ENDPOINT}}#$ETCD_ENDPOINT#g" /etc/cloudone/configuration.json

cd /src/cloudone
./cloudone &

while :
do
	sleep 1
done

