FROM mcr.microsoft.com/devcontainers/go:1-1.24-bookworm

ARG TZ
ENV TZ="$TZ"

# Install basic development tools and iptables/ipset
RUN apt update && apt install -y less \
  git \
  procps \
  sudo \
  fzf \
  zsh \
  man-db \
  unzip \
  gnupg2 \
  gh \
  iptables \
  ipset \
  iproute2 \
  dnsutils \
  aggregate \
  jq \
  curl \
  wget

# Install Node.js for Claude Code
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - && \
  apt-get install -y nodejs

# Set up vscode user (matches devcontainer defaults)
ARG USERNAME=vscode

# Create directories and set permissions (as root)
RUN mkdir -p /commandhistory /workspace && \
  touch /commandhistory/.bash_history && \
  chown -R $USERNAME:$USERNAME /commandhistory /workspace

# Set `DEVCONTAINER` environment variable to help with orientation
ENV DEVCONTAINER=true

USER $USERNAME

WORKDIR /workspace

# Install delta for git diff
RUN ARCH=$(dpkg --print-architecture) && \
  wget "https://github.com/dandavison/delta/releases/download/0.18.2/git-delta_0.18.2_${ARCH}.deb" && \
  sudo dpkg -i "git-delta_0.18.2_${ARCH}.deb" && \
  rm "git-delta_0.18.2_${ARCH}.deb"

# Set the default shell to zsh
ENV SHELL=/bin/zsh

# Default powerline10k theme
RUN sh -c "$(wget -O- https://github.com/deluan/zsh-in-docker/releases/download/v1.2.0/zsh-in-docker.sh)" -- \
  -p git \
  -p fzf \
  -a "source /usr/share/doc/fzf/examples/key-bindings.zsh" \
  -a "source /usr/share/doc/fzf/examples/completion.zsh" \
  -a "export PROMPT_COMMAND='history -a' && export HISTFILE=/commandhistory/.bash_history" \
  -x

# Install Claude Code
RUN sudo npm install -g @anthropic-ai/claude-code



# Set permissions for workspace and config (final ownership fix)
RUN chown -R vscode:vscode /workspace /commandhistory

USER vscode