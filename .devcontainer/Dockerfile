FROM mcr.microsoft.com/devcontainers/go:1.24

# Install additional development tools
RUN go install golang.org/x/tools/gopls@latest && \
    go install github.com/go-delve/delve/cmd/dlv@latest && \
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest && \
    go install golang.org/x/tools/cmd/goimports@latest && \
    go install github.com/cweill/gotests/gotests@latest && \
    go install github.com/fatih/gomodifytags@latest && \
    go install github.com/josharian/impl@latest && \
    go install github.com/haya14busa/goplay/cmd/goplay@latest && \
    go install github.com/ramya-rao-a/go-outline@latest && \
    go install github.com/rogpeppe/godef@latest && \
    go install github.com/uudashr/gopkgs/v2/cmd/gopkgs@latest && \
    go install github.com/zdjones/bolt@latest

# Set up shell environment
RUN echo 'export PATH=$PATH:/home/vscode/go/bin' >> /home/vscode/.bashrc && \
    echo 'export PATH=$PATH:/home/vscode/go/bin' >> /home/vscode/.zshrc 