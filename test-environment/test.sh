#!/bin/bash
# Test script for agena

# Clean up
rm agena/demo-task/*.log
rm .fixed-item-*

# Check for --inactivity-test flag
if [[ "$1" == "--inactivity-test" ]]; then
    # Simulate long delays to test the 30-second inactivity timer
    echo "Testing inactivity timer (will pause for 35s between messages)..."
    MOCK_GEMINI_INACTIVITY_TEST=1 agena demo-task
else
    # Normal test run
    agena demo-task
fi
