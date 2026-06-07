FROM dbox-base:latest

# Ruby と Bundler のインストール
RUN apt-get update && apt-get install -y \
    ruby-full \
    ruby-bundler \
    && rm -rf /var/lib/apt/lists/*
