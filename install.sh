#!/bin/bash
set -e

# Check if Go is installed
if ! command -v go &> /dev/null; then
  echo "Go is not installed."
  if [[ "$OSTYPE" == "darwin"* ]]; then
    echo "Install with: brew install go"
  else
    echo "Install with: sudo apt install golang  (Debian/Ubuntu)"
    echo "         or:  sudo dnf install golang   (Fedora)"
    echo "         or:  visit https://go.dev/dl"
  fi
  exit 1
fi

echo "Go found: $(go version)"

# Clone and run
INSTALL_DIR="$HOME/cc-lens"

if [ -d "$INSTALL_DIR" ]; then
  echo "Updating existing install..."
  cd "$INSTALL_DIR" && git pull
else
  git clone https://github.com/SemihMutlu07/cc-lens.git "$INSTALL_DIR"
  cd "$INSTALL_DIR"
fi

echo "Starting CC Lens..."
go run . &
sleep 2

# Open browser
if command -v xdg-open &> /dev/null; then
  xdg-open http://localhost:8080
elif command -v open &> /dev/null; then
  open http://localhost:8080
else
  echo "Open http://localhost:8080 in your browser."
fi

echo "CC Lens is running at http://localhost:8080"
