FROM ubuntu:22.04

# Gerekli sistem araçlarını yüklüyoruz
RUN apt update && \
    apt install -y python3 python3-pip tmux bash && \
    apt clean

# Çalışma dizinini tanımla
WORKDIR /workspace

# Varsayılan komut (terminal açılınca bash'e düşsün)
CMD ["/bin/bash"]
