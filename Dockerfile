# latest kali base image
FROM kalilinux/kali-last-release@sha256:7f275b601a53f1bee1bd96b1d229128b282cfe9d1d71a31226978575e8693484

LABEL "project"="gr3ysh3ll"
LABEL "author"="fr3d"
LABEL "version"="v1.6.1"

ENV DEBIAN_FRONTEND=noninteractive \
    TZ="America/New_York"

# Bootstrap HTTPS support and install base utilities
RUN echo "deb http://kali.download/kali kali-rolling main contrib non-free non-free-firmware" > /etc/apt/sources.list && \
  apt-get update && \
  apt-get install -y --no-install-recommends ca-certificates && \
  echo "deb https://kali.download/kali kali-rolling main contrib non-free non-free-firmware" > /etc/apt/sources.list && \
  apt-get update && \
  apt-get install -y --no-install-recommends \
    curl \
    git \
    nano \
    sudo \
    tmux \
    vim \
    wget \
    zsh && \
  apt-get clean && rm -rf /var/lib/apt/lists/*

RUN groupadd --gid 1000 kali \
  && useradd --home-dir /home/kali --create-home --uid 1000 \
    --gid 1000 --shell /usr/bin/zsh --skel /dev/null kali \
  && chown -R kali:kali /home/kali/ \
  && echo kali:kali | chpasswd \
  && usermod -aG sudo kali \
  && echo 'kali  ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers.d/kali

WORKDIR /home/kali/

USER kali

COPY --chown=kali:kali sources/ /tmp/sources/

RUN sudo chmod +x /tmp/sources/*.sh && /tmp/sources/0-base.sh

RUN /tmp/sources/1-tools.sh

#uncomment for bug bounty usage
#RUN /tmp/sources/2-tools.sh

COPY --chown=kali:kali resources /home/kali/resources/

RUN /tmp/sources/3-home.sh && \
  sudo rm -rf /tmp/sources /home/kali/resources

CMD ["/usr/bin/zsh"]
