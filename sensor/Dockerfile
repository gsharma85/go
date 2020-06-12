FROM golang:1.14

# Set up github ssh keys
RUN mkdir /root/.ssh
WORKDIR /root/.ssh
COPY id_rsa .
COPY passfile .
ADD ssh-add-pass.sh .
RUN chmod 700 id_rsa
RUN chmod 700 passfile
RUN chmod +x ssh-add-pass.sh

RUN apt-get update -y
RUN apt-get install -y expect

RUN	ssh-keyscan -H github.com >> /etc/ssh/ssh_known_hosts

# set up GO private repo
RUN go env -w GOPRIVATE=github.com/gsharma85/go/
RUN echo "[url \"git@github.com:\"]\n\tinsteadOf = https://github.com/" >> /root/.gitconfig
RUN echo "StrictHostKeyChecking no" > /root/.ssh/config
RUN mkdir /root/gotest
WORKDIR /
RUN /root/.ssh/ssh-add-pass.sh /root/.ssh/id_rsa /root/.ssh/passfile  
    #ssh -T git@github.com
    #git clone git@github.com:gsharma85/go.git


CMD ["/go/bin/sensor", "-d" , "/root/gotest"]