#!/bin/bash

# Exit on error
set -e

# Configuration
LLAMA_REPO="https://github.com/ggerganov/llama.cpp.git"
LLAMA_BRANCH="master"
INSTALL_DIR="$HOME/.passiver"
MODEL_DIR="$INSTALL_DIR/models"
LLAMA_DIR="$INSTALL_DIR/llama.cpp"
BUILD_DIR="$LLAMA_DIR/build"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}Installing Passiver AI Node...${NC}"

# Create installation directory
mkdir -p "$INSTALL_DIR"
mkdir -p "$MODEL_DIR"

# Install system dependencies
echo -e "${GREEN}Installing system dependencies...${NC}"
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    # Linux
    sudo apt-get update
    sudo apt-get install -y \
        build-essential \
        cmake \
        git \
        python3 \
        python3-pip \
        cuda-toolkit \
        ninja-build
elif [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    if ! command -v brew &> /dev/null; then
        /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
    fi
    brew install cmake python git ninja
fi

# Clone and build llama.cpp
echo -e "${GREEN}Cloning and building llama.cpp...${NC}"
if [ ! -d "$LLAMA_DIR" ]; then
    git clone --branch $LLAMA_BRANCH $LLAMA_REPO "$LLAMA_DIR"
fi

cd "$LLAMA_DIR"
git pull

# Create build directory
mkdir -p "$BUILD_DIR"
cd "$BUILD_DIR"

# Configure and build with CMake
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    # Linux with CUDA support
    cmake .. -DLLAMA_CUBLAS=ON -DCMAKE_CUDA_ARCHITECTURES=all-major
    cmake --build . --config Release -j
else
    # macOS or other platforms (Metal support on macOS)
    if [[ "$OSTYPE" == "darwin"* ]]; then
        cmake .. -DLLAMA_METAL=ON
    else
        cmake ..
    fi
    cmake --build . --config Release -j
fi

# Create model download script
cat > "$LLAMA_DIR/scripts/download-model.sh" << 'EOF'
#!/bin/bash

MODEL_NAME=$1
MODEL_DIR=$2

case $MODEL_NAME in
    "llama-7b-q4")
        URL="https://huggingface.co/TheBloke/Llama-2-7B-GGUF/resolve/main/llama-2-7b.Q4_K_M.gguf"
        ;;
    "llama-13b-q4")
        URL="https://huggingface.co/TheBloke/Llama-2-13B-GGUF/resolve/main/llama-2-13b.Q4_K_M.gguf"
        ;;
    *)
        echo "Unknown model: $MODEL_NAME"
        exit 1
        ;;
esac

echo "Downloading $MODEL_NAME..."
curl -L $URL -o "$MODEL_DIR/$MODEL_NAME.gguf"
EOF

chmod +x "$LLAMA_DIR/scripts/download-model.sh"

# Download default model
echo -e "${GREEN}Downloading default model...${NC}"
"$LLAMA_DIR/scripts/download-model.sh" "llama-7b-q4" "$MODEL_DIR"

# Create configuration
cat > "$INSTALL_DIR/config.json" << EOF
{
    "model_path": "$MODEL_DIR",
    "llama_path": "$BUILD_DIR",
    "default_model": "llama-7b-q4.gguf"
}
EOF

echo -e "${GREEN}Installation complete!${NC}"
echo "Installation directory: $INSTALL_DIR"
echo "Model directory: $MODEL_DIR"
echo "llama.cpp directory: $LLAMA_DIR"
echo "Build directory: $BUILD_DIR" 