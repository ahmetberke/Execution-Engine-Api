# Ubuntu 22.04 tabanlı özel image oluşturuyoruz
FROM ubuntu:22.04

# Sistem paketlerini güncelliyoruz
RUN apt update && apt upgrade -y

# Python ve gerekli bağımlılıkları yüklüyoruz
RUN apt install -y python3 python3-pip

RUN apt update && apt install -y tmux bash

# Varsayılan shell olarak bash kullanıyoruz
CMD ["/bin/bash"]