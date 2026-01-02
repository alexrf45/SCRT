# latest kali base image
FROM kalilinux/kali-rolling:latest

LABEL "project"="gr3ysh3ll"
LABEL "author"="fr3d"
LABEL "version"="v1.6.1"

ENV DEBIAN_FRONTEND=noninteractive
ENV TZ="America/New_York"

RUN apt-get update && apt-get install -y sudo \
  tmux \
  vim \
  nano \
  wget \
  curl \
  git \
  zsh

RUN groupadd --gid 1000 kali \
  && useradd --home-dir /home/kali --create-home --uid 1000 \
  --gid 1000 --shell /bin/bash --skel /dev/null kali

RUN chown -R kali:kali /home/kali/ \
  && echo kali:kali | chpasswd \
  && usermod -aG sudo kali \
  && echo 'kali  ALL=(ALL) NOPASSWD:ALL' >>  /etc/sudoers.d/kali

WORKDIR /home/kali/

USER kali

COPY sources/ /tmp/sources

RUN sudo chown -R kali:kali /tmp/sources/* && sudo chmod +x /tmp/sources/*.sh

RUN /tmp/sources/0-base.sh

RUN /tmp/sources/1-tools.sh

#uncomment for bug bounty usage
#RUN  /tmp/sources/2-tools.sh

COPY resources /home/kali/resources/

RUN sudo chown -R kali:kali /home/kali/resources && /tmp/sources/3-home.sh

RUN sudo rm -rf /tmp/sources && sudo rm -rf /home/kali/resources

USER kali

RUN sudo chsh $USER -s /usr/bin/zsh

CMD ["/usr/bin/zsh"]



