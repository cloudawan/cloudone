FROM ubuntu:14.04

# Install wget
RUN apt-get install -y wget

# Install for Docker
RUN apt-get install -y apt-transport-https
# Install docker with the specific version
RUN echo deb http://get.docker.com/ubuntu docker main > /etc/apt/sources.list.d/docker.list
RUN apt-key adv --keyserver pgp.mit.edu --recv-keys 36A1D7869245C8950F966E92D8576A8BA88D21E9
RUN apt-get update
RUN apt-get install -y lxc-docker-1.6.1

# Install git
RUN apt-get install -y git

# Install curl
RUN apt-get install -y curl

# Install python
RUN apt-get install -y python

# Download pip
RUN apt-get install -y python-pip

# Install python httplib2
RUN pip install httplib2

# Bundle app source
COPY . /src

# Copy configuration
RUN mkdir -p /etc/cloudone
RUN mkdir -p /var/log/cloudone
RUN cp /src/cloudone/configuration/* /etc/cloudone/

# Privilege
RUN chmod a+x /src/run.sh
RUN chmod a+x /src/cloudone/cloudone

EXPOSE  8081
CMD ["/src/run.sh"]
