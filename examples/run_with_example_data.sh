#!/bin/bash

# Script to run BubbleCal with example data for screenshots
# This temporarily symlinks the example .bubblecal folder to your home directory

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
EXAMPLE_DIR="$SCRIPT_DIR/.bubblecal"
HOME_BUBBLECAL="$HOME/.bubblecal"
BACKUP_DIR="$HOME/.bubblecal.backup"

# Check if bubblecal binary exists
if [ ! -f "../bubblecal" ]; then
    echo "Building bubblecal..."
    (cd .. && go build -o bubblecal ./cmd/bubblecal/)
fi

# Backup existing .bubblecal if it exists
if [ -e "$HOME_BUBBLECAL" ]; then
    echo "Backing up existing ~/.bubblecal to ~/.bubblecal.backup"
    mv "$HOME_BUBBLECAL" "$BACKUP_DIR"
fi

# Create symlink to example data
echo "Using example data from $EXAMPLE_DIR"
ln -s "$EXAMPLE_DIR" "$HOME_BUBBLECAL"

# Run bubblecal
echo "Running BubbleCal with example data..."
echo "Press 'q' to quit when done taking screenshots"
../bubblecal

# Clean up
echo "Cleaning up..."
rm "$HOME_BUBBLECAL"

# Restore backup if it exists
if [ -e "$BACKUP_DIR" ]; then
    echo "Restoring original ~/.bubblecal"
    mv "$BACKUP_DIR" "$HOME_BUBBLECAL"
fi

echo "Done!"