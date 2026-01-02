#!/usr/bin/env python3
"""
Security Research Container Toolkit (SCRT)

A CLI tool for managing Docker-based security research environments.
Provides project isolation, standardized tooling, and persistent workspaces.

Version: 2.0.0
Author: fr3d
"""

import os
import sys
import json
import shutil
import tarfile
import argparse
import subprocess
import re
from pathlib import Path
from datetime import datetime
from typing import Optional, List, Dict, Tuple
from dataclasses import dataclass, field, asdict

from rich.console import Console
from rich.table import Table
from rich.prompt import Prompt, Confirm
from rich.progress import Progress, SpinnerColumn, TextColumn
import requests


# ============================================================================
# CONSTANTS
# ============================================================================

DEFAULT_IMAGE = "fonalex45/scrt:latest"
DEFAULT_SHELL = "/bin/zsh"
DEFAULT_BACKUP_DIR = "./backups"
DOCKER_HUB_API = "https://hub.docker.com/v2/repositories/fonalex45/scrt/tags"
DOCKER_HUB_URL = "https://hub.docker.com/r/fonalex45/scrt/tags"

# Project directory structure
PROJECT_DIRECTORIES = [
    'recon',
    'www',
    'exploit',
    'pivot',
    'privesc',
    'report',
    '.gr3ysh3ll-logs'
]

# Default Linux capabilities for container
DEFAULT_CAPABILITIES = ["NET_ADMIN", "CAP_SYS_TIME"]

# Docker label for identifying SCRT containers
CONTAINER_LABEL = "author=fr3d"

# Project name validation pattern
PROJECT_NAME_PATTERN = r'^[a-zA-Z0-9_-]+$'


# ============================================================================
# CONSOLE COLORS
# ============================================================================

class Colors:
    """ANSI color codes for rich console output."""
    ERROR = "red"
    SUCCESS = "green"
    WARNING = "yellow"
    INFO = "blue"
    HEADER = "bold cyan"
    ACCENT = "magenta"


# ============================================================================
# GLOBAL CONSOLE
# ============================================================================

console = Console()


# ============================================================================
# CONFIGURATION
# ============================================================================

@dataclass
class Config:
    """
    Configuration settings for SCRT.

    Attributes:
        docker_image: Docker image to use for containers
        container_shell: Default shell inside containers
        host_networking: Enable host network mode
        enable_x11: Enable X11 forwarding for GUI apps
        enable_gpu: Enable GPU passthrough
        custom_caps: Linux capabilities to add to containers
        extra_mounts: Additional volume mounts
        work_dir_base: Base directory for project workspaces
        config_file: Path to configuration file
    """
    docker_image: str = DEFAULT_IMAGE
    container_shell: str = DEFAULT_SHELL
    host_networking: bool = True
    enable_x11: bool = True
    enable_gpu: bool = True
    custom_caps: List[str] = field(
        default_factory=lambda: DEFAULT_CAPABILITIES.copy())
    extra_mounts: List[str] = field(default_factory=list)
    work_dir_base: str = field(default_factory=lambda: os.getcwd())
    config_file: Path = field(
        default_factory=lambda: Path.home() / ".scrt.conf.json")

    def load(self) -> None:
        """Load configuration from JSON file if it exists."""
        if not self.config_file.exists():
            return

        try:
            with open(self.config_file, 'r') as f:
                data = json.load(f)
                self._update_from_dict(data)
            console.print(f"[{Colors.INFO}]Configuration loaded from {
                          self.config_file}[/]")
        except Exception as e:
            console.print(f"[{Colors.WARNING}]Failed to load config: {e}[/]")

    def save(self) -> None:
        """Save current configuration to JSON file."""
        try:
            with open(self.config_file, 'w') as f:
                json.dump(asdict(self), f, indent=2, default=str)
            console.print(f"[{Colors.SUCCESS}]Configuration saved to {
                          self.config_file}[/]")
        except Exception as e:
            console.print(f"[{Colors.ERROR}]Failed to save config: {e}[/]")

    def _update_from_dict(self, data: dict) -> None:
        """Update configuration attributes from dictionary."""
        for key, value in data.items():
            if hasattr(self, key):
                setattr(self, key, value)


# ============================================================================
# DOCKER MANAGER
# ============================================================================

class DockerManager:
    """
    Manages all Docker operations for SCRT containers.

    Handles container lifecycle, image management, and Docker API interactions.
    """

    def __init__(self, config: Config):
        """
        Initialize DockerManager with configuration.

        Args:
            config: SCRT configuration object
        """
        self.config = config

    # ------------------------------------------------------------------------
    # Docker Environment Checks
    # ------------------------------------------------------------------------

    def check_docker(self) -> bool:
        """
        Verify Docker is installed and running.

        Returns:
            True if Docker daemon is accessible, False otherwise
        """
        try:
            result = subprocess.run(
                ['docker', 'info'],
                capture_output=True,
                text=True,
                timeout=5
            )
            return result.returncode == 0
        except (subprocess.SubprocessError, FileNotFoundError):
            return False

    # ------------------------------------------------------------------------
    # Container State Checks
    # ------------------------------------------------------------------------

    def container_exists(self, project: str) -> bool:
        """
        Check if a container with the given project name exists.

        Args:
            project: Project/container name

        Returns:
            True if container exists (running or stopped)
        """
        try:
            result = subprocess.run(
                ['docker', 'container', 'inspect', project],
                capture_output=True,
                text=True
            )
            return result.returncode == 0
        except subprocess.SubprocessError:
            return False

    def container_running(self, project: str) -> bool:
        """
        Check if a container is currently running.

        Args:
            project: Project/container name

        Returns:
            True if container is running
        """
        try:
            result = subprocess.run(
                ['docker', 'container', 'inspect', '-f',
                    '{{.State.Running}}', project],
                capture_output=True,
                text=True
            )
            return result.stdout.strip() == 'true'
        except subprocess.SubprocessError:
            return False

    # ------------------------------------------------------------------------
    # Image Management
    # ------------------------------------------------------------------------

    def get_available_tags(self) -> Tuple[List[str], Optional[str]]:
        """
        Fetch available Docker image tags from Docker Hub.

        Returns:
            Tuple of (sorted tag list, most recent version tag)
        """
        try:
            response = requests.get(DOCKER_HUB_API, timeout=5)

            if response.status_code != 200:
                return self._get_default_tags()

            data = response.json()
            tags = [tag['name'] for tag in data.get('results', [])]

            return self._sort_tags(tags)

        except Exception as e:
            console.print(
                f"[{Colors.WARNING}]Could not fetch tags from Docker Hub: {e}[/]")
            return self._get_default_tags()

    def pull_image(self, image: str) -> bool:
        """
        Pull a Docker image from registry.

        Args:
            image: Full image name (repository:tag)

        Returns:
            True if pull succeeded
        """
        try:
            with Progress(
                SpinnerColumn(),
                TextColumn("[progress.description]{task.description}"),
                console=console
            ) as progress:
                task = progress.add_task(
                    f"Pulling image {image}...", total=None)

                result = subprocess.run(
                    ['docker', 'pull', image],
                    capture_output=True,
                    text=True
                )

                progress.update(task, completed=True)

            if result.returncode == 0:
                console.print(
                    f"[{Colors.SUCCESS}]Successfully pulled {image}[/]")
                return True
            else:
                console.print(f"[{Colors.ERROR}]Failed to pull image: {
                              result.stderr}[/]")
                return False

        except subprocess.SubprocessError as e:
            console.print(f"[{Colors.ERROR}]Error pulling image: {e}[/]")
            return False

    # ------------------------------------------------------------------------
    # Container Lifecycle
    # ------------------------------------------------------------------------

    def start_container(self, project: str, image: Optional[str] = None) -> bool:
        """
        Create and start a new SCRT container.

        Args:
            project: Project name (becomes container name)
            image: Docker image to use (defaults to config)

        Returns:
            True if container started successfully
        """
        if not self.validate_project_name(project):
            return False

        if self._handle_existing_container(project):
            return False

        # Setup project workspace
        if not self._create_project_structure(project):
            return False

        # Build and execute Docker run command
        image = image or self.config.docker_image
        cmd = self._build_docker_run_command(project, image)

        return self._execute_docker_run(cmd, project)

    def enter_container(self, project: str) -> bool:
        """
        Enter an existing container with an interactive shell.

        Args:
            project: Project/container name

        Returns:
            True if successfully entered container
        """
        if not self.container_exists(project):
            console.print(f"[{Colors.ERROR}]Container '{
                          project}' does not exist[/]")
            return False

        # Start container if it's stopped
        if not self.container_running(project):
            console.print(f"[{Colors.INFO}]Starting stopped container '{
                          project}'...[/]")
            subprocess.run(
                ['docker', 'container', 'start', project],
                capture_output=True
            )

        console.print(f"[{Colors.INFO}]Entering container '{project}'...[/]")

        try:
            subprocess.run([
                'docker', 'exec', '-it', project,
                self.config.container_shell
            ])
            return True
        except subprocess.SubprocessError as e:
            console.print(f"[{Colors.ERROR}]Failed to enter container: {e}[/]")
            return False

    def stop_container(self, project: str) -> bool:
        """
        Stop a running container.

        Args:
            project: Project/container name

        Returns:
            True if container stopped successfully
        """
        if not self.container_exists(project):
            console.print(f"[{Colors.WARNING}]Container '{
                          project}' does not exist[/]")
            return False

        if not self.container_running(project):
            console.print(f"[{Colors.INFO}]Container '{
                          project}' is already stopped[/]")
            return True

        console.print(f"[{Colors.INFO}]Stopping container '{project}'...[/]")

        try:
            subprocess.run(
                ['docker', 'container', 'stop', project],
                capture_output=True
            )
            console.print(f"[{Colors.SUCCESS}]Container '{
                          project}' stopped[/]")
            return True
        except subprocess.SubprocessError as e:
            console.print(f"[{Colors.ERROR}]Failed to stop container: {e}[/]")
            return False

    def destroy_container(self, project: str, force: bool = False) -> bool:
        """
        Remove container and optionally delete project data.

        Args:
            project: Project/container name
            force: Skip confirmation prompts

        Returns:
            True if destruction succeeded
        """
        project_dir = Path(self.config.work_dir_base) / project

        # Handle case where container doesn't exist
        if not self.container_exists(project):
            return self._cleanup_orphaned_directory(project_dir, force)

        # Confirm destruction
        if not force and not self._confirm_destruction(project):
            return False

        return self._remove_container_and_data(project, project_dir)

    # ------------------------------------------------------------------------
    # Data Management
    # ------------------------------------------------------------------------

    def backup_project(self, project: str, backup_dir: str = DEFAULT_BACKUP_DIR) -> bool:
        """
        Create a compressed backup of project data.

        Args:
            project: Project name
            backup_dir: Directory to store backup file

        Returns:
            True if backup created successfully
        """
        project_path = Path(self.config.work_dir_base) / project

        if not project_path.exists():
            console.print(f"[{Colors.ERROR}]Project directory '{
                          project_path}' does not exist[/]")
            return False

        backup_file = self._create_backup_path(project, backup_dir)
        console.print(f"[{Colors.INFO}]Creating backup of '{project}'...[/]")

        try:
            with tarfile.open(backup_file, 'w:gz') as tar:
                tar.add(project_path, arcname=project)

            console.print(f"[{Colors.SUCCESS}]Backup created: {
                          backup_file}[/]")
            return True

        except Exception as e:
            console.print(f"[{Colors.ERROR}]Failed to create backup: {e}[/]")
            return False

    # ------------------------------------------------------------------------
    # Container Listing
    # ------------------------------------------------------------------------

    def list_containers(self) -> List[Dict[str, str]]:
        """
        List all SCRT containers (running and stopped).

        Returns:
            List of container information dictionaries
        """
        try:
            result = subprocess.run(
                [
                    'docker', 'ps', '-a',
                    '--filter', f'label={CONTAINER_LABEL}',
                    '--format', 'json'
                ],
                capture_output=True,
                text=True
            )

            return self._parse_container_list(result.stdout)

        except subprocess.SubprocessError:
            return []

    # ------------------------------------------------------------------------
    # Validation
    # ------------------------------------------------------------------------

    @staticmethod
    def validate_project_name(project: str) -> bool:
        """
        Validate project name against allowed pattern.

        Args:
            project: Proposed project name

        Returns:
            True if valid
        """
        if not re.match(PROJECT_NAME_PATTERN, project):
            console.print(
                f"[{Colors.ERROR}]Invalid project name. "
                f"Use only alphanumeric characters, hyphens, and underscores[/]"
            )
            return False
        return True

    # ------------------------------------------------------------------------
    # Private Helper Methods
    # ------------------------------------------------------------------------

    def _get_default_tags(self) -> Tuple[List[str], Optional[str]]:
        """Return default tags when API fetch fails."""
        return ['latest', 'dev'], None

    def _sort_tags(self, tags: List[str]) -> Tuple[List[str], Optional[str]]:
        """Sort tags with priority ordering (latest, dev, versions)."""
        priority_tags = []
        version_tags = []

        for tag in tags:
            if tag in ['latest', 'dev']:
                priority_tags.append(tag)
            elif tag.startswith('v'):
                version_tags.append(tag)

        # Sort version tags in reverse order
        version_tags.sort(key=lambda x: x.lstrip('v'), reverse=True)
        most_recent_version = version_tags[0] if version_tags else None

        return priority_tags + version_tags, most_recent_version

    def _handle_existing_container(self, project: str) -> bool:
        """Check for existing container and inform user."""
        if not self.container_exists(project):
            return False

        console.print(f"[{Colors.WARNING}]Container '{
                      project}' already exists[/]")

        if self.container_running(project):
            console.print(
                f"[{Colors.INFO}]Container is running. Use 'enter' to access it.[/]")
        else:
            console.print(
                f"[{Colors.INFO}]Container is stopped. Use 'enter' to start and access it.[/]")

        return True

    def _create_project_structure(self, project: str) -> bool:
        """Create project directory structure."""
        project_dir = Path(self.config.work_dir_base) / project

        try:
            for dir_name in PROJECT_DIRECTORIES:
                (project_dir / dir_name).mkdir(parents=True, exist_ok=True)

            console.print(
                f"[{Colors.INFO}]Created project structure for '{project}'[/]")
            return True

        except OSError as e:
            console.print(
                f"[{Colors.ERROR}]Failed to create project structure: {e}[/]")
            return False

    def _build_docker_run_command(self, project: str, image: str) -> List[str]:
        """Build complete Docker run command with all options."""
        cmd = ['docker', 'run', '--name', project, '-it']

        # Network configuration
        if self.config.host_networking:
            cmd.append('--net=host')

        # Linux capabilities
        for cap in self.config.custom_caps:
            cmd.append(f'--cap-add={cap}')

        # GPU support
        if self.config.enable_gpu and Path('/dev/dri').exists():
            cmd.append('--device=/dev/dri:/dev/dri')

        # X11 forwarding
        if self.config.enable_x11:
            self._add_x11_options(cmd)

        # Environment variables
        self._add_environment_variables(cmd, project)

        # Volume mounts
        self._add_volume_mounts(cmd, project)

        # Working directory and entrypoint
        cmd.extend(['-w', f'/{project}'])
        cmd.extend(['--entrypoint', self.config.container_shell])
        cmd.append(image)

        return cmd

    def _add_x11_options(self, cmd: List[str]) -> None:
        """Add X11 forwarding options to Docker command."""
        display = os.environ.get('DISPLAY')
        if not display:
            return

        cmd.extend(['-e', f'DISPLAY={display}'])
        cmd.extend(['-v', '/tmp/.X11-unix:/tmp/.X11-unix'])

        xauth = Path.home() / '.Xauthority'
        if xauth.exists():
            cmd.extend(['-v', f'{xauth}:{xauth}'])

    def _add_environment_variables(self, cmd: List[str], project: str) -> None:
        """Add environment variables to Docker command."""
        cmd.extend(['-e', f'TARGET={project}'])
        cmd.extend(['-e', f'TZ={os.environ.get("TZ", "UTC")}'])

    def _add_volume_mounts(self, cmd: List[str], project: str) -> None:
        """Add volume mounts to Docker command."""
        project_dir = Path(self.config.work_dir_base) / project

        # Standard mounts
        cmd.extend(['-v', f'{project_dir}/.gr3ysh3ll-logs:/root/.logs:rw'])
        cmd.extend(['-v', f'{project_dir}:/{project}'])

        # Extra custom mounts
        for mount in self.config.extra_mounts:
            cmd.extend(['-v', mount])

    def _execute_docker_run(self, cmd: List[str], project: str) -> bool:
        """Execute Docker run command."""
        console.print(f"[{Colors.INFO}]Starting container '{project}'...[/]")

        try:
            subprocess.run(cmd)
            return True
        except subprocess.SubprocessError as e:
            console.print(f"[{Colors.ERROR}]Failed to start container: {e}[/]")
            return False

    def _cleanup_orphaned_directory(self, project_dir: Path, force: bool) -> bool:
        """Clean up project directory when container doesn't exist."""
        if not project_dir.exists():
            console.print(
                f"[{Colors.WARNING}]Container and directory do not exist[/]")
            return True

        if not force and not Confirm.ask(f"Remove project directory '{project_dir}'?"):
            return False

        shutil.rmtree(project_dir)
        console.print(f"[{Colors.SUCCESS}]Project directory removed[/]")
        return True

    def _confirm_destruction(self, project: str) -> bool:
        """Confirm container destruction with user."""
        return Confirm.ask(
            f"[{Colors.WARNING}]Destroy container '{
                project}' and all its data?[/]"
        )

    def _remove_container_and_data(self, project: str, project_dir: Path) -> bool:
        """Remove container and associated project data."""
        try:
            # Remove container
            subprocess.run(
                ['docker', 'container', 'rm', '-f', project],
                capture_output=True
            )

            # Remove project directory
            if project_dir.exists():
                shutil.rmtree(project_dir)

            console.print(f"[{Colors.SUCCESS}]Container '{
                          project}' destroyed[/]")
            return True

        except (subprocess.SubprocessError, OSError) as e:
            console.print(
                f"[{Colors.ERROR}]Failed to destroy container: {e}[/]")
            return False

    def _create_backup_path(self, project: str, backup_dir: str) -> Path:
        """Create backup file path with timestamp."""
        backup_path = Path(backup_dir)
        backup_path.mkdir(parents=True, exist_ok=True)

        timestamp = datetime.now().strftime('%Y-%m-%d_%H-%M-%S')
        return backup_path / f"{timestamp}_{project}.tar.gz"

    def _parse_container_list(self, stdout: str) -> List[Dict[str, str]]:
        """Parse Docker ps JSON output into container info list."""
        containers = []

        for line in stdout.strip().split('\n'):
            if not line:
                continue

            try:
                data = json.loads(line)
                containers.append({
                    'name': data.get('Names', ''),
                    'status': data.get('Status', ''),
                    'image': data.get('Image', ''),
                    'createdat': data.get('CreatedAt', '')
                })
            except json.JSONDecodeError:
                continue

        return containers


# ============================================================================
# COMMAND LINE INTERFACE
# ============================================================================

class CLI:
    """
    Command-line interface for SCRT.

    Handles argument parsing, user interaction, and command execution.
    """

    def __init__(self):
        """Initialize CLI with configuration and Docker manager."""
        self.config = Config()
        self.config.load()
        self.docker_manager = DockerManager(self.config)

    # ------------------------------------------------------------------------
    # User Interface
    # ------------------------------------------------------------------------

    def show_banner(self) -> None:
        """Display ASCII art banner."""
        banner = """
[bold cyan]
             ________________________________________________
            /                                                \\
           |    _________________________________________     |
           |   |                                         |    |
           |   |  admin@scrt $ _                         |    |
           |   |  Security Research Container Toolkit    |    |
           |   |                   (SCRT)                |    |
           |   |_________________________________________|    |
           |                                                  |
            \\_________________________________________________/
                   \\___________________________________/
[/]
        """
        console.print(banner)

    def select_docker_tag(self) -> str:
        """
        Interactive Docker tag selection menu.

        Returns:
            Full Docker image name with selected tag
        """
        console.print(f"\n[{Colors.HEADER}]Docker Tag Selection[/]")
        console.print(f"[{Colors.INFO}]View available tags: {
                      DOCKER_HUB_URL}[/]\n")

        tags, most_recent = self.docker_manager.get_available_tags()

        # Display tag options
        table = self._create_tag_selection_table(most_recent)
        console.print(table)

        # Get user selection
        max_choice = "4" if most_recent else "3"
        choice = Prompt.ask("\nSelect an option", choices=[
                            "1", "2", "3", "4"][:int(max_choice)])

        return self._process_tag_selection(choice, most_recent)

    # ------------------------------------------------------------------------
    # Main CLI Entry Point
    # ------------------------------------------------------------------------

    def run_cli(self) -> None:
        """Parse arguments and execute requested command."""
        parser = self._create_argument_parser()
        args = parser.parse_args()

        # Show help if no command provided
        if not args.command:
            self.show_banner()
            parser.print_help()
            return

        # Verify Docker availability (except for config command)
        if args.command != 'config' and not self.docker_manager.check_docker():
            console.print(
                f"[{Colors.ERROR}]Docker is not available or not running[/]")
            sys.exit(1)

        # Route to appropriate command handler
        self._execute_command(args)

    # ------------------------------------------------------------------------
    # Configuration Management
    # ------------------------------------------------------------------------

    def manage_config(self) -> None:
        """Interactive configuration editor."""
        console.print(f"\n[{Colors.HEADER}]Configuration Management[/]")

        # Display current configuration
        self._display_current_config()

        # Prompt for modifications
        if not Confirm.ask("\nModify configuration?"):
            return

        self._update_config_interactively()
        self.config.save()

    # ------------------------------------------------------------------------
    # Private Helper Methods
    # ------------------------------------------------------------------------

    def _create_tag_selection_table(self, most_recent: Optional[str]) -> Table:
        """Create formatted table for tag selection."""
        table = Table(
            title="Available Tags",
            show_header=True,
            header_style="bold magenta"
        )
        table.add_column("Option", style="cyan", width=10)
        table.add_column("Tag", style="green")
        table.add_column("Description", style="yellow")

        table.add_row("1", "latest", "Latest stable release")
        table.add_row("2", "dev", "Development version")

        if most_recent:
            table.add_row("3", most_recent, "Most recent version tag")
            table.add_row("4", "custom", "Enter custom tag")
        else:
            table.add_row("3", "custom", "Enter custom tag")

        return table

    def _process_tag_selection(self, choice: str, most_recent: Optional[str]) -> str:
        """Process user's tag selection and return full image name."""
        if choice == "1":
            return "fonalex45/scrt:latest"
        elif choice == "2":
            return "fonalex45/scrt:dev"
        elif choice == "3" and most_recent:
            return f"fonalex45/scrt:{most_recent}"
        else:
            custom_tag = Prompt.ask("Enter custom tag")
            if not custom_tag.startswith("fonalex45/scrt:"):
                custom_tag = f"fonalex45/scrt:{custom_tag}"
            return custom_tag

    def _create_argument_parser(self) -> argparse.ArgumentParser:
        """Create and configure argument parser."""
        parser = argparse.ArgumentParser(
            description='Security Research Container Toolkit',
            formatter_class=argparse.RawDescriptionHelpFormatter
        )

        subparsers = parser.add_subparsers(dest='command', help='Commands')

        # Define all subcommands
        self._add_start_command(subparsers)
        self._add_enter_command(subparsers)
        self._add_stop_command(subparsers)
        self._add_destroy_command(subparsers)
        self._add_backup_command(subparsers)
        self._add_pull_command(subparsers)
        self._add_list_command(subparsers)
        self._add_config_command(subparsers)

        return parser

    def _add_start_command(self, subparsers) -> None:
        """Add 'start' subcommand."""
        start = subparsers.add_parser('start', help='Start a new container')
        start.add_argument('project', help='Project name')
        start.add_argument('--image', help='Docker image to use')
        start.add_argument('--select-tag', action='store_true',
                           help='Interactively select Docker tag')

    def _add_enter_command(self, subparsers) -> None:
        """Add 'enter' subcommand."""
        enter = subparsers.add_parser('enter', help='Enter a container')
        enter.add_argument('project', help='Project name')

    def _add_stop_command(self, subparsers) -> None:
        """Add 'stop' subcommand."""
        stop = subparsers.add_parser('stop', help='Stop a container')
        stop.add_argument('project', help='Project name')

    def _add_destroy_command(self, subparsers) -> None:
        """Add 'destroy' subcommand."""
        destroy = subparsers.add_parser('destroy', help='Destroy a container')
        destroy.add_argument('project', help='Project name')
        destroy.add_argument('--force', action='store_true',
                             help='Force destroy without confirmation')

    def _add_backup_command(self, subparsers) -> None:
        """Add 'backup' subcommand."""
        backup = subparsers.add_parser('backup', help='Backup project data')
        backup.add_argument('project', help='Project name')
        backup.add_argument('--dir', default=DEFAULT_BACKUP_DIR,
                            help='Backup directory')

    def _add_pull_command(self, subparsers) -> None:
        """Add 'pull' subcommand."""
        pull = subparsers.add_parser('pull', help='Pull SCRT image')
        pull.add_argument('--image', help='SCRT image to pull')
        pull.add_argument('--select-tag', action='store_true',
                          help='Interactively select Docker tag')

    def _add_list_command(self, subparsers) -> None:
        """Add 'list' subcommand."""
        subparsers.add_parser('list', help='List all containers')

    def _add_config_command(self, subparsers) -> None:
        """Add 'config' subcommand."""
        subparsers.add_parser('config', help='Manage configuration')

    def _execute_command(self, args: argparse.Namespace) -> None:
        """Route parsed arguments to appropriate command handler."""
        command_map = {
            'start': self._handle_start,
            'enter': self._handle_enter,
            'stop': self._handle_stop,
            'destroy': self._handle_destroy,
            'backup': self._handle_backup,
            'pull': self._handle_pull,
            'list': self._handle_list,
            'config': self._handle_config
        }

        handler = command_map.get(args.command)
        if handler:
            handler(args)

    def _handle_start(self, args: argparse.Namespace) -> None:
        """Handle 'start' command."""
        image = args.image
        if args.select_tag:
            image = self.select_docker_tag()
        self.docker_manager.start_container(args.project, image)

    def _handle_enter(self, args: argparse.Namespace) -> None:
        """Handle 'enter' command."""
        self.docker_manager.enter_container(args.project)

    def _handle_stop(self, args: argparse.Namespace) -> None:
        """Handle 'stop' command."""
        self.docker_manager.stop_container(args.project)

    def _handle_destroy(self, args: argparse.Namespace) -> None:
        """Handle 'destroy' command."""
        self.docker_manager.destroy_container(args.project, args.force)

    def _handle_backup(self, args: argparse.Namespace) -> None:
        """Handle 'backup' command."""
        self.docker_manager.backup_project(args.project, args.dir)

    def _handle_pull(self, args: argparse.Namespace) -> None:
        """Handle 'pull' command."""
        image = args.image
        if args.select_tag or not image:
            image = self.select_docker_tag()
        self.docker_manager.pull_image(image)

    def _handle_list(self, args: argparse.Namespace) -> None:
        """Handle 'list' command."""
        containers = self.docker_manager.list_containers()

        if not containers:
            console.print(f"[{Colors.INFO}]No containers found[/]")
            return

        table = Table(title="SCRT Containers", show_header=True)
        table.add_column("Name", style="cyan")
        table.add_column("Status", style="yellow")
        table.add_column("Image", style="green")
        table.add_column("CreatedAt", style="red")

        for container in containers:
            table.add_row(
                container['name'],
                container['status'],
                container['image'],
                container['createdat']
            )

        console.print(table)

    def _handle_config(self, args: argparse.Namespace) -> None:
        """Handle 'config' command."""
        self.manage_config()

    def _display_current_config(self) -> None:
        """Display current configuration in a table."""
        table = Table(title="Current Configuration", show_header=True)
        table.add_column("Setting", style="cyan")
        table.add_column("Value", style="green")

        for key, value in asdict(self.config).items():
            if key != 'config_file':
                table.add_row(key, str(value))

        console.print(table)

    def _update_config_interactively(self) -> None:
        """Prompt user to update configuration settings."""
        self.config.docker_image = Prompt.ask(
            "Docker image",
            default=self.config.docker_image
        )

        self.config.container_shell = Prompt.ask(
            "Container shell",
            default=self.config.container_shell
        )

        self.config.host_networking = Confirm.ask(
            "Enable host networking?",
            default=self.config.host_networking
        )

        self.config.enable_x11 = Confirm.ask(
            "Enable X11 forwarding?",
            default=self.config.enable_x11
        )

        self.config.enable_gpu = Confirm.ask(
            "Enable GPU support?",
            default=self.config.enable_gpu
        )

        caps = Prompt.ask(
            "Custom capabilities (comma-separated)",
            default=','.join(self.config.custom_caps)
        )
        self.config.custom_caps = [cap.strip() for cap in caps.split(',')]

        self.config.work_dir_base = Prompt.ask(
            "Work directory base",
            default=self.config.work_dir_base
        )


def main() -> None:
    """
    Main entry point for SCRT CLI.
    Handles initialization, error catching, and graceful shutdown.
    """
    cli = CLI()
    if len(sys.argv) == 1:
        cli.show_banner()
        console.print(f"\n[{Colors.INFO}]Run with --help for CLI usage[/]")
    else:
        cli.run_cli()


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        console.print(f"\n[{Colors.WARNING}]Operation cancelled by user[/]")
        sys.exit(0)
    except Exception as e:
        console.print(f"[{Colors.ERROR}]Unexpected error: {e}[/]")
        sys.exit(1)
