FROM docker/sandbox-templates:opencode-docker

# nvim のインストール（sbx ベースイメージには含まれていない）
RUN sudo apt-get update && sudo apt-get install -y \
    neovim \
    && sudo rm -rf /var/lib/apt/lists/*

# nvim 設定を agent ユーザーにコピー
COPY nvim/ /home/agent/.config/nvim/
RUN sudo chown -R agent:agent /home/agent/.config/nvim/

WORKDIR /workspace
