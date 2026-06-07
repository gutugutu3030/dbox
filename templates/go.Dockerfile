FROM dbox-base:latest

# Go のインストール
RUN curl -fsSL https://go.dev/dl/go1.24.0.linux-amd64.tar.gz | tar -C /usr/local -xz

ENV PATH=$PATH:/usr/local/go/bin
ENV GOPATH=/root/go
