FROM docker/sandbox-templates:opencode-docker

# nvim のインストール（sbx ベースイメージには含まれていない）
RUN sudo apt-get update && sudo apt-get install -y \
    neovim \
    && sudo rm -rf /var/lib/apt/lists/*

WORKDIR /workspace
