# Python Package Development Guide for gr3ysh3ll

## What is pyproject.toml?

`pyproject.toml` is the modern standard for Python project configuration. It replaces the old `setup.py` approach and centralizes all project metadata, dependencies, and tool configurations in one file.

## Project Setup Steps

### 1. Create the Project Structure

```bash
mkdir gr3ysh3ll
cd gr3ysh3ll

# Create the directory structure
mkdir -p src/gr3ysh3ll/{core,utils,config}
mkdir -p tests
mkdir -p docs

# Create essential files
touch src/gr3ysh3ll/__init__.py
touch src/gr3ysh3ll/__main__.py
touch src/gr3ysh3ll/cli.py
touch README.md
touch LICENSE
```

### 2. Set Up Virtual Environment

```bash
# Create virtual environment
python -m venv venv

# Activate it (Linux/Mac)
source venv/bin/activate
# Or on Windows:
# venv\Scripts\activate

# Upgrade pip
pip install --upgrade pip
```

### 3. Install Your Package in Development Mode

```bash
# Install package in editable mode with dev dependencies
pip install -e ".[dev]"

# This installs:
# - Your package in development mode
# - All runtime dependencies (click, docker, rich, etc.)
# - All development dependencies (pytest, black, mypy, etc.)
```

## Understanding the pyproject.toml Sections

### [build-system]
Specifies how to build your package. Uses modern setuptools.

### [project]
Core project metadata:
- **name**: Package name (what users will `pip install`)
- **version**: Current version (update this for releases)
- **dependencies**: Runtime requirements
- **scripts**: Command-line entry points

### [project.optional-dependencies]
Optional dependency groups:
- **dev**: Development tools (testing, linting, formatting)
- **docs**: Documentation building tools
- **test**: Testing-specific dependencies

### [project.scripts]
Creates command-line tools:
```toml
gr3ysh3ll = "gr3ysh3ll.cli:main"  # Creates `gr3ysh3ll` command
gr3y = "gr3ysh3ll.cli:main"       # Creates `gr3y` alias
```

## Development Workflow

### 1. Install Dependencies

```bash
# Install just runtime dependencies
pip install -e .

# Install with development tools
pip install -e ".[dev]"

# Install with docs tools
pip install -e ".[dev,docs]"
```

### 2. Code Formatting and Linting

```bash
# Format code with black
black src/ tests/

# Sort imports with isort
isort src/ tests/

# Lint with flake8
flake8 src/ tests/

# Type checking with mypy
mypy src/
```

### 3. Testing

```bash
# Run all tests
pytest

# Run with coverage
pytest --cov=gr3ysh3ll --cov-report=html

# Run only unit tests
pytest -m unit

# Run only integration tests
pytest -m integration
```

### 4. Building and Distribution

```bash
# Install build tools
pip install build twine

# Build the package
python -m build

# This creates:
# - dist/gr3ysh3ll-1.0.0.tar.gz (source distribution)
# - dist/gr3ysh3ll-1.0.0-py3-none-any.whl (wheel)
```

### 5. Publishing (when ready)

```bash
# Test upload to TestPyPI first
twine upload --repository testpypi dist/*

# Upload to PyPI
twine upload dist/*
```

## Essential Files to Create

### src/gr3ysh3ll/__init__.py
```python
"""gr3ysh3ll - Optimized Python wrapper for penetration testing containers."""

__version__ = "1.0.0"
__author__ = "Your Name"
__email__ = "your.email@example.com"
```

### src/gr3ysh3ll/cli.py
```python
"""Command-line interface for gr3ysh3ll."""

import click
from rich.console import Console

console = Console()

@click.group()
@click.version_option()
def main():
    """gr3ysh3ll - Penetration testing container manager."""
    pass

@main.command()
@click.argument('project')
def start(project):
    """Start a new container."""
    console.print(f"Starting container: {project}", style="green")

if __name__ == "__main__":
    main()
```

### src/gr3ysh3ll/__main__.py
```python
"""Allow running gr3ysh3ll as a module with python -m gr3ysh3ll."""

from .cli import main

if __name__ == "__main__":
    main()
```

## Usage After Installation

Once installed, users can:

```bash
# Use the main command
gr3ysh3ll start myproject

# Use the alias
gr3y start myproject

# Run as module
python -m gr3ysh3ll start myproject
```

## Version Management

Update the version in `pyproject.toml` for each release:

```toml
[project]
version = "1.0.1"  # Update this
```

## Common Commands

```bash
# Development installation
pip install -e ".[dev]"

# Run tests
pytest

# Format code
black src/ tests/
isort src/ tests/

# Type check
mypy src/

# Build package
python -m build

# Clean build artifacts
rm -rf build/ dist/ src/*.egg-info/
```

## Tips

1. **Always use a virtual environment** to avoid conflicts
2. **Pin major versions** in dependencies to avoid breaking changes
3. **Use semantic versioning** (MAJOR.MINOR.PATCH)
4. **Test before publishing** using TestPyPI
5. **Keep README.md updated** with installation and usage instructions
6. **Add a LICENSE file** for open source projects

This modern approach makes your package easy to install, develop, and distribute while following Python packaging best practices!