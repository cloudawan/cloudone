#!/bin/bash

# Check whether Cassandra is up
# Parameter
retry_amount=120
cassandra_host_list_with_quotation=${CASSANDRA_CLUSTER_HOST//,}
cassandra_host_list=${cassandra_host_list_with_quotation//\"}
cassandra_port=$CASSANDRA_CLUSTER_PORT

is_cassandra_up() {
  for cassandra_host in $cassandra_host_list
  do
    cassandra_url="http://$cassandra_host:$cassandra_port"
    cassandra_response=$(curl -m 1 "$cassandra_url")

    if [[ $cassandra_response == *"Invalid or unsupported protocol"* ]]; then
      return 1
    fi
  done
  return 0
}

# Cassandra
for ((i=0;i<$retry_amount;i++))
do
  echo "ping $i times to Cassandra"
  is_cassandra_up
  cassandra_result=$?
  if [ $cassandra_result == 1 ]; then
	break
  fi
  sleep 1
done

if [ $i == $retry_amount ]; then
  echo "Could not get ping response from Cassandra"
  exit -1
fi






# Use environment
sed -i "s/{{KUBEAPI_HOST}}/$KUBEAPI_HOST/g" /etc/cloudone/configuration.json
sed -i "s/{{KUBEAPI_PORT}}/$KUBEAPI_PORT/g" /etc/cloudone/configuration.json
sed -i "s/{{CASSANDRA_CLUSTER_HOST}}/$CASSANDRA_CLUSTER_HOST/g" /etc/cloudone/configuration.json
sed -i "s/{{CASSANDRA_CLUSTER_PORT}}/$CASSANDRA_CLUSTER_PORT/g" /etc/cloudone/configuration.json
sed -i "s/{{CASSANDRA_CLUSTER_REPLICATION_STRATEGY}}/$CASSANDRA_CLUSTER_REPLICATION_STRATEGY/g" /etc/cloudone/configuration.json
sed -i "s/{{EMAIL_SENDER_ACCOUNT}}/$EMAIL_SENDER_ACCOUNT/g" /etc/cloudone/configuration.json
sed -i "s/{{EMAIL_SENDER_PASSWORD}}/$EMAIL_SENDER_PASSWORD/g" /etc/cloudone/configuration.json
sed -i "s/{{EMAIL_SENDER_HOST}}/$EMAIL_SENDER_HOST/g" /etc/cloudone/configuration.json
sed -i "s/{{EMAIL_SENDER_PORT}}/$EMAIL_SENDER_PORT/g" /etc/cloudone/configuration.json
# Use different delimiter since URL contains slash
sed -i "s#{{SMS_NEXMO_URL}}#$SMS_NEXMO_URL#g" /etc/cloudone/configuration.json
sed -i "s/{{SMS_NEXMO_API_KEY}}/$SMS_NEXMO_API_KEY/g" /etc/cloudone/configuration.json
sed -i "s/{{SMS_NEXMO_API_SECRET}}/$SMS_NEXMO_API_SECRET/g" /etc/cloudone/configuration.json
sed -i "s/{{GLUSTERFS_HOST}}/$GLUSTERFS_HOST/g" /etc/cloudone/configuration.json
# Use different delimiter since URL contains slash
sed -i "s#{{GLUSTERFS_MOUNT_PATH}}#$GLUSTERFS_MOUNT_PATH#g" /etc/cloudone/configuration.json
sed -i "s/{{GLUSTERFS_SSH_USER}}/$GLUSTERFS_SSH_USER/g" /etc/cloudone/configuration.json
sed -i "s/{{GLUSTERFS_SSH_PASSWORD}}/$GLUSTERFS_SSH_PASSWORD/g" /etc/cloudone/configuration.json

cd /src/cloudone
./cloudone &

while :
do
	sleep 1
done

