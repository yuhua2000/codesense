#!/bin/bash

# Basic usage examples for CodeSense

echo "=== CodeSense Examples ==="
echo ""

# Example 1: Analyze current directory
echo "Example 1: Analyze current directory with GPT-4"
echo "codesense  -model gpt-4-turbo -lang en"
echo ""

# Example 2: Analyze specific project
echo "Example 2: Analyze a Go project and save to file"
echo "codesense  -path ~/go/src/github.com/user/project -model gpt-4-turbo -output analysis.md"
echo ""

# Example 3: Use environment variable for API key
echo "Example 3: Using environment variable"
echo "export OPENAI_API_KEY='your-api-key'"
echo "codesense  -path . -model gpt-3.5-turbo"
echo ""

# Example 4: Custom API endpoint
echo "Example 4: Using custom OpenAI-compatible endpoint"
echo "codesense  -path . -model gpt-4-turbo -url 'https://api.openai.com/v1/' -key 'sk-...'"
echo ""

# Example 5: Chinese output
echo "Example 5: Generate Chinese report"
echo "codesense  -path . -model gpt-4-turbo -lang zh -output 分析报告.md"
echo ""

echo "=== Tips ==="
echo "1. Start with GPT-3.5-turbo for faster/cheaper analysis"
echo "2. Use -output flag to save reports for later reference"
echo "3. For large projects, the AI may need to read multiple files"
echo "4. Check the generated report for insights about your codebase"