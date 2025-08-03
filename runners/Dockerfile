# GitHub Actions Runner with Prerequisites
# Based on official GitHub Actions runner
FROM ghcr.io/actions/actions-runner:2.319.1

# Switch to root for package installation
USER root

# Set environment variables
ENV DEBIAN_FRONTEND=noninteractive
ENV RUNNER_MANUALLY_TRAP_SIG=1
ENV ACTIONS_RUNNER_PRINT_LOG_TO_STDOUT=1

# Update package list and install essential packages
RUN apt-get update && apt-get install -y \
    # Basic utilities
    curl \
    wget \
    unzip \
    zip \
    git \
    jq \
    tree \
    htop \
    vim \
    nano \
    # Build tools
    build-essential \
    make \
    gcc \
    g++ \
    # Development languages
    python3 \
    python3-pip \
    python3-venv \
    nodejs \
    npm \
    # System tools
    ca-certificates \
    software-properties-common \
    apt-transport-https \
    lsb-release \
    gnupg \
    # Network tools
    netcat-openbsd \
    telnet \
    iputils-ping \
    # Archive tools
    tar \
    gzip \
    bzip2 \
    xz-utils \
    && rm -rf /var/lib/apt/lists/*

# Install Go (latest stable version)
RUN curl -fsSL https://golang.org/dl/go1.21.5.linux-amd64.tar.gz | tar -C /usr/local -xzf - \
    && ln -s /usr/local/go/bin/go /usr/local/bin/go \
    && ln -s /usr/local/go/bin/gofmt /usr/local/bin/gofmt

# Install Azure CLI
RUN curl -sL https://aka.ms/InstallAzureCLIDeb | bash

# Install kubectl
RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" \
    && chmod +x kubectl \
    && mv kubectl /usr/local/bin/

# Install Helm
RUN curl https://baltocdn.com/helm/signing.asc | gpg --dearmor | tee /usr/share/keyrings/helm.gpg > /dev/null \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/helm.gpg] https://baltocdn.com/helm/stable/debian/ all main" | tee /etc/apt/sources.list.d/helm-stable-debian.list \
    && apt-get update \
    && apt-get install helm \
    && rm -rf /var/lib/apt/lists/*

# Install Terraform
RUN wget -O- https://apt.releases.hashicorp.com/gpg | gpg --dearmor | tee /usr/share/keyrings/hashicorp-archive-keyring.gpg \
    && echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" | tee /etc/apt/sources.list.d/hashicorp.list \
    && apt-get update \
    && apt-get install terraform \
    && rm -rf /var/lib/apt/lists/*

# Install Docker Compose
RUN curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose \
    && chmod +x /usr/local/bin/docker-compose

# Install Python packages commonly used in CI/CD
RUN pip3 install --no-cache-dir \
    awscli \
    azure-cli \
    requests \
    pyyaml \
    pytest \
    black \
    flake8 \
    mypy \
    coverage

# Install Node.js tools
RUN npm install -g \
    yarn \
    pnpm \
    typescript \
    eslint \
    prettier \
    @vue/cli \
    @angular/cli

# Install additional development tools
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /usr/local/bin v1.54.2

# Install GitHub CLI
RUN curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg \
    && chmod go+r /usr/share/keyrings/githubcli-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | tee /etc/apt/sources.list.d/github-cli.list > /dev/null \
    && apt-get update \
    && apt-get install gh -y \
    && rm -rf /var/lib/apt/lists/*

# Create directories for custom tools
RUN mkdir -p /usr/local/scripts

# Copy any custom scripts (if they exist)
COPY runners/scripts/ /usr/local/scripts/
RUN chmod +x /usr/local/scripts/* || true

# Set up environment variables
ENV PATH="/usr/local/go/bin:/usr/local/scripts:${PATH}"
ENV GOPATH="/home/runner/go"
ENV GOROOT="/usr/local/go"

# Create Go workspace for runner user
RUN mkdir -p /home/runner/go/{bin,src,pkg} \
    && chown -R runner:runner /home/runner/go

# Switch back to runner user
USER runner

# Set working directory
WORKDIR /home/runner

# Create common directories
RUN mkdir -p ~/.local/bin \
    && mkdir -p ~/.cache \
    && mkdir -p ~/.config

# Add local bin to PATH for runner user
ENV PATH="/home/runner/.local/bin:${PATH}"

# Install user-level Python packages
RUN pip3 install --user \
    pipenv \
    poetry \
    virtualenv

# Set default shell to bash for better compatibility
SHELL ["/bin/bash", "-c"]

# Display installed versions (for verification)
RUN echo "=== Installed Tools ===" \
    && echo "Go: $(go version)" \
    && echo "Node: $(node --version)" \
    && echo "npm: $(npm --version)" \
    && echo "Python: $(python3 --version)" \
    && echo "pip: $(pip3 --version)" \
    && echo "Azure CLI: $(az --version | head -1)" \
    && echo "kubectl: $(kubectl version --client --short)" \
    && echo "Helm: $(helm version --short)" \
    && echo "Terraform: $(terraform --version | head -1)" \
    && echo "Docker Compose: $(docker-compose --version)" \
    && echo "GitHub CLI: $(gh --version | head -1)" \
    && echo "golangci-lint: $(golangci-lint --version)" \
    && echo "======================"
