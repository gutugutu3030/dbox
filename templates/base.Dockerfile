FROM ubuntu:24.04

# 基本ツールのインストール
RUN apt-get update && apt-get install -y \
    neovim \
    git \
    curl \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# nvim 設定をコピー（ビルドコンテキストに nvim/ ディレクトリが必要）
COPY nvim/ /root/.config/nvim/

WORKDIR /workspace
