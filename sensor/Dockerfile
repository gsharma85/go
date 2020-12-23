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
WORKDIR /
RUN CGO_ENABLED=0 GOOS=linux /root/.ssh/go-get.sh /root/.ssh/id_rsa /root/.ssh/passfile

FROM alpine:latest 
RUN mkdir /root/gotest
RUN mkdir /root/sensor
WORKDIR /root/sensor
COPY actorsystem.json .

WORKDIR /
RUN mkdir /go/
RUN mkdir /go/bin

COPY --from=0 /go/bin/sensor /go/bin/
CMD ["/go/bin/sensor", "-cf" , "/root/sensor/actorsystem.json", "-lf", "/root/sensor/sensor.log"]