FROM mcr.microsoft.com/devcontainers/go:1.23

# Install additional OS packages
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends \
    curl \
    git \
    make \
    && apt-get clean -y \
    && rm -rf /var/lib/apt/lists/*

# Set up non-root user
USER vscode
WORKDIR /workspace

# Install Go tools
RUN go install -v \
    golang.org/x/tools/gopls@latest \
    github.com/go-delve/delve/cmd/dlv@latest \
    honnef.co/go/tools/cmd/staticcheck@latest