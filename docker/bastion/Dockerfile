FROM alpine:3.7

# add bash and openssh
RUN apk add --update --no-cache bash openssh

# change root password, then allow root login via ssh
RUN mkdir /var/run/sshd
RUN echo 'root:iamverysecret' | chpasswd
RUN sed -i 's/PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config

## generate hostkeys
#RUN ssh-keygen -A

# copy host keys over
COPY ./ssh /etc/ssh

# copy ssh keys and authorized_keys file, then set permissions
COPY ./id_docker_bastion /root/.ssh/id_docker_bastion
COPY ./id_docker_bastion.pub /root/.ssh/id_docker_bastion.pub
COPY ./authorized_keys /root/.ssh/authorized_keys
RUN chmod 700 /root/.ssh
RUN chmod 600 /root/.ssh/id_docker_bastion
RUN chmod 644 /root/.ssh/id_docker_bastion.pub
RUN chmod 644 /root/.ssh/authorized_keys

# expose port 22
EXPOSE 22

# when container starts it will start the ssh daemon
CMD ["/usr/sbin/sshd", "-D"]