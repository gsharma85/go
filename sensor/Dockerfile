FROM golang:1.14

# Set up github ssh keys
RUN mkdir /root/.ssh
WORKDIR /root/.ssh
COPY id_rsa .
COPY passfile .
ADD go-get.sh .
RUN chmod 700 id_rsa
RUN chmod 700 passfile
RUN chmod +x go-get.sh

RUN apt-get update -y
RUN apt-get install -y expect

RUN	ssh-keyscan -H github.com >> /etc/ssh/ssh_known_hosts

# set up GO private repo
RUN go env -w GOPRIVATE=github.com/gsharma85/go/
RUN echo "[url \"git@github.com:\"]\n\tinsteadOf = https://github.com/" >> /root/.gitconfig
RUN echo "StrictHostKeyChecking no" > /root/.ssh/config
RUN mkdir /root/gotest
WORKDIR /
RUN /root/.ssh/go-get.sh /root/.ssh/id_rsa /root/.ssh/passfile


CMD ["/go/bin/sensor", "-d" , "/root/gotest", "-lf", "/root/sensor.log"]