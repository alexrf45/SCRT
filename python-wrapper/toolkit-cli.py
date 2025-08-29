#!/usr/bin/env python3
"""
toolkit
"""

import os
import sys
import json
import shutil
import tarfile
import argparse
import subprocess
from pathlib import Path
from datetime import datetime
from typing import Optional, List, Dict, Tuple
from dataclasses import dataclass, field, asdict

# Check for required packages
try:
    import rich
    from rich.console import Console
    from rich.table import Table
    from rich.prompt import Prompt, Confirm
    from rich.progress import Progress, SpinnerColumn, TextColumn
    from rich.text import Text
    from rich.layout import Layout
except ImportError:
    print("Error: Required packages not installed.")
    print("Please install: pip install rich requests")
    sys.exit(1)

try:
    import requests
except ImportError:
    print("Error: requests package not installed.")
    print("Please install: pip install requests")
    sys.exit(1)

console = Console()

# Color scheme


class Colors:
    ERROR = "red"
    SUCCESS = "green"
    WARNING = "yellow"
    INFO = "blue"
    HEADER = "bold cyan"
    ACCENT = "magenta"


@dataclass
class Config:
    """Configuration for the toolkit"""
    docker_image: str = "fonalex45/toolkit:latest"
    container_shell: str = "/bin/zsh"
    host_networking: bool = True
    enable_x11: bool = True
    enable_gpu: bool = True
    custom_caps: List[str] = field(default_factory=lambda: [
                                   "NET_ADMIN", "CAP_SYS_TIME"])
    extra_mounts: List[str] = field(default_factory=list)
    work_dir_base: str = field(default_factory=lambda: os.getcwd())
    config_file: Path = field(
        default_factory=lambda: Path.home() / ".toolkit.conf.json")

    def load(self):
        """Load configuration from file"""
        if self.config_file.exists():
            try:
                with open(self.config_file, 'r') as f:
                    data = json.load(f)
                    for key, value in data.items():
                        if hasattr(self, key):
                            setattr(self, key, value)
                console.print(f"[{Colors.INFO}]Configuration loaded from {
                              self.config_file}[/]")
            except Exception as e:
                console.print(
                    f"[{Colors.WARNING}]Failed to load config: {e}[/]")

    def save(self):
        """Save configuration to file"""
        try:
            with open(self.config_file, 'w') as f:
                json.dump(asdict(self), f, indent=2, default=str)
            console.print(f"[{Colors.SUCCESS}]Configuration saved to {
                          self.config_file}[/]")
        except Exception as e:
            console.print(f"[{Colors.ERROR}]Failed to save config: {e}[/]")


class DockerManager:
    """Manages Docker operations"""

    def __init__(self, config: Config):
        self.config = config

    def check_docker(self) -> bool:
        """Check if Docker is available and running"""
        try:
            result = subprocess.run(['docker', 'info'],
                                    capture_output=True,
                                    text=True,
                                    timeout=5)
            return result.returncode == 0
        except (subprocess.SubprocessError, FileNotFoundError):
            return False

    def container_exists(self, project: str) -> bool:
        """Check if container exists"""
        try:
            result = subprocess.run(['docker', 'container', 'inspect', project],
                                    capture_output=True,
                                    text=True)
            return result.returncode == 0
        except subprocess.SubprocessError:
            return False

    def container_running(self, project: str) -> bool:
        """Check if container is running"""
        try:
            result = subprocess.run(['docker', 'container', 'inspect', '-f',
                                     '{{.State.Running}}', project],
                                    capture_output=True,
                                    text=True)
            return result.stdout.strip() == 'true'
        except subprocess.SubprocessError:
            return False

    def get_available_tags(self) -> Tuple[List[str], Optional[str]]:
        """Fetch available Docker tags from Docker Hub"""
        try:
            # Get tags from Docker Hub API
            url = "https://hub.docker.com/v2/repositories/fonalex45/toolkit/tags"
            response = requests.get(url, timeout=5)

            if response.status_code == 200:
                data = response.json()
                tags = [tag['name'] for tag in data.get('results', [])]

                # Sort tags, putting 'latest' and 'dev' first
                priority_tags = []
                version_tags = []

                for tag in tags:
                    if tag in ['latest', 'dev']:
                        priority_tags.append(tag)
                    elif tag.startswith('v'):
                        version_tags.append(tag)

                # Sort version tags by version number
                version_tags.sort(key=lambda x: x.lstrip('v'), reverse=True)
                most_recent_version = version_tags[0] if version_tags else None

                return priority_tags + version_tags, most_recent_version
            else:
                return ['latest', 'dev'], None

        except Exception as e:
            console.print(
                f"[{Colors.WARNING}]Could not fetch tags from Docker Hub: {e}[/]")
            return ['latest', 'dev'], None

    def pull_image(self, image: str) -> bool:
        """Pull Docker image"""
        try:
            with Progress(
                SpinnerColumn(),
                TextColumn("[progress.description]{task.description}"),
                console=console
            ) as progress:
                task = progress.add_task(
                    f"Pulling image {image}...", total=None)

                result = subprocess.run(['docker', 'pull', image],
                                        capture_output=True,
                                        text=True)

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

    def start_container(self, project: str, image: Optional[str] = None) -> bool:
        """Start a new container"""
        if not self.validate_project_name(project):
            return False

        if self.container_exists(project):
            console.print(f"[{Colors.WARNING}]Container '{
                          project}' already exists[/]")
            if self.container_running(project):
                console.print(
                    f"[{Colors.INFO}]Container is running. Use 'enter' to access it.[/]")
            else:
                console.print(
                    f"[{Colors.INFO}]Container is stopped. Use 'enter' to start and access it.[/]")
            return False

        # Create project directory structure
        project_dir = Path(self.config.work_dir_base) / project
        dirs = ['recon', 'www', 'exploit', 'pivot',
                'privesc', 'report', '.gr3ysh3ll-logs']

        for dir_name in dirs:
            (project_dir / dir_name).mkdir(parents=True, exist_ok=True)

        console.print(
            f"[{Colors.INFO}]Created project structure for '{project}'[/]")

        # Build Docker command
        image = image or self.config.docker_image
        cmd = ['docker', 'run', '--name', project, '-it']

        # Networking
        if self.config.host_networking:
            cmd.append('--net=host')

        # Capabilities
        for cap in self.config.custom_caps:
            cmd.append(f'--cap-add={cap}')

        # GPU support
        if self.config.enable_gpu and Path('/dev/dri').exists():
            cmd.append('--device=/dev/dri:/dev/dri')

        # X11 forwarding
        if self.config.enable_x11 and os.environ.get('DISPLAY'):
            cmd.extend(['-e', f'DISPLAY={os.environ["DISPLAY"]}'])
            cmd.extend(['-v', '/tmp/.X11-unix:/tmp/.X11-unix'])
            xauth = Path.home() / '.Xauthority'
            if xauth.exists():
                cmd.extend(['-v', f'{xauth}:{xauth}'])

        # Environment variables
        cmd.extend(['-e', f'TARGET={project}'])
        cmd.extend(['-e', f'TZ={os.environ.get("TZ", "UTC")}'])

        # Mounts
        cmd.extend(['-v', f'{project_dir}/.gr3ysh3ll-logs:/root/.logs:rw'])
        cmd.extend(['-v', f'{project_dir}:/{project}'])

        # Extra mounts
        for mount in self.config.extra_mounts:
            cmd.extend(['-v', mount])

        # Working directory and entrypoint
        cmd.extend(['-w', f'/{project}'])
        cmd.extend(['--entrypoint', self.config.container_shell])
        cmd.append(image)

        console.print(f"[{Colors.INFO}]Starting container '{project}'...[/]")

        try:
            subprocess.run(cmd)
            return True
        except subprocess.SubprocessError as e:
            console.print(f"[{Colors.ERROR}]Failed to start container: {e}[/]")
            return False

    def enter_container(self, project: str) -> bool:
        """Enter an existing container"""
        if not self.container_exists(project):
            console.print(f"[{Colors.ERROR}]Container '{
                          project}' does not exist[/]")
            return False

        if not self.container_running(project):
            console.print(f"[{Colors.INFO}]Starting stopped container '{
                          project}'...[/]")
            subprocess.run(['docker', 'container', 'start', project],
                           capture_output=True)

        console.print(f"[{Colors.INFO}]Entering container '{project}'...[/]")

        try:
            subprocess.run(['docker', 'exec', '-it', project,
                           self.config.container_shell])
            return True
        except subprocess.SubprocessError as e:
            console.print(f"[{Colors.ERROR}]Failed to enter container: {e}[/]")
            return False

    def stop_container(self, project: str) -> bool:
        """Stop a running container"""
        if not self.container_exists(project):
            console.print(f"[{Colors.WARNING}]Container '{
                          project}' does not exist[/]")
            return False

        if self.container_running(project):
            console.print(f"[{Colors.INFO}]Stopping container '{
                          project}'...[/]")
            try:
                subprocess.run(['docker', 'container', 'stop', project],
                               capture_output=True)
                console.print(f"[{Colors.SUCCESS}]Container '{
                              project}' stopped[/]")
                return True
            except subprocess.SubprocessError as e:
                console.print(
                    f"[{Colors.ERROR}]Failed to stop container: {e}[/]")
                return False
        else:
            console.print(f"[{Colors.INFO}]Container '{
                          project}' is already stopped[/]")
            return True

    def destroy_container(self, project: str, force: bool = False) -> bool:
        """Destroy container and its data"""
        if not self.container_exists(project):
            console.print(f"[{Colors.WARNING}]Container '{
                          project}' does not exist[/]")

            # Clean up directory if it exists
            project_dir = Path(self.config.work_dir_base) / project
            if project_dir.exists():
                if not force:
                    if not Confirm.ask(f"Remove project directory '{project_dir}'?"):
                        return False
                shutil.rmtree(project_dir)
                console.print(
                    f"[{Colors.SUCCESS}]Project directory removed[/]")
            return True

        if not force:
            if not Confirm.ask(f"[{Colors.WARNING}]Destroy container '{project}' and all its data?[/]"):
                console.print("[{Colors.INFO}]Operation cancelled[/]")
                return False

        try:
            subprocess.run(['docker', 'container', 'rm', '-f', project],
                           capture_output=True)

            project_dir = Path(self.config.work_dir_base) / project
            if project_dir.exists():
                shutil.rmtree(project_dir)

            console.print(f"[{Colors.SUCCESS}]Container '{
                          project}' destroyed[/]")
            return True

        except (subprocess.SubprocessError, OSError) as e:
            console.print(
                f"[{Colors.ERROR}]Failed to destroy container: {e}[/]")
            return False

    def backup_project(self, project: str, backup_dir: str = "./backups") -> bool:
        """Backup project data"""
        project_path = Path(self.config.work_dir_base) / project

        if not project_path.exists():
            console.print(f"[{Colors.ERROR}]Project directory '{
                          project_path}' does not exist[/]")
            return False

        backup_path = Path(backup_dir)
        backup_path.mkdir(parents=True, exist_ok=True)

        timestamp = datetime.now().strftime('%Y-%m-%d_%H-%M-%S')
        backup_file = backup_path / f"{timestamp}_{project}.tar.gz"

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

    def list_containers(self) -> List[Dict[str, str]]:
        """List all toolkit containers"""
        try:
            result = subprocess.run([
                'docker', 'ps', '-a',
                '--filter', 'ancestor=fonalex45/toolkit',
                '--format', 'json'
            ], capture_output=True, text=True)

            containers = []
            for line in result.stdout.strip().split('\n'):
                if line:
                    try:
                        data = json.loads(line)
                        containers.append({
                            'name': data.get('Names', ''),
                            'status': data.get('Status', ''),
                            'image': data.get('Image', '')
                        })
                    except json.JSONDecodeError:
                        continue

            return containers

        except subprocess.SubprocessError:
            return []

    @staticmethod
    def validate_project_name(project: str) -> bool:
        """Validate project name"""
        import re
        if not re.match(r'^[a-zA-Z0-9_-]+$', project):
            console.print(
                f"[{Colors.ERROR}]Invalid project name. Use only alphanumeric characters, hyphens, and underscores[/]")
            return False
        return True


class CLI:
    """Command Line Interface"""

    def __init__(self):
        self.config = Config()
        self.config.load()
        self.docker_manager = DockerManager(self.config)

    def show_banner(self):
        """Display ASCII banner"""
        banner = """
[bold cyan]
             ________________________________________________
            /                                                \\
           |    _________________________________________     |
           |   |                                         |    |
           |   |  admin@toolkit $ _                      |    |
           |   |  Security Research Container Manager    |    |
           |   |                                         |    |
           |   |_________________________________________|    |
           |                                                  |
            \\_________________________________________________/
                   \\___________________________________/
[/]
        """
        console.print(banner)

    def select_docker_tag(self) -> str:
        """Interactive Docker tag selection"""
        console.print(f"\n[{Colors.HEADER}]Docker Tag Selection[/]")
        console.print(
            f"[{Colors.INFO}]View available tags: https://hub.docker.com/r/fonalex45/toolkit/tags[/]\n")

        tags, most_recent = self.docker_manager.get_available_tags()

        # Create options table
        table = Table(title="Available Tags", show_header=True,
                      header_style="bold magenta")
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

        console.print(table)

        choice = Prompt.ask("\nSelect an option", choices=[
                            "1", "2", "3", "4"] if most_recent else ["1", "2", "3"])

        if choice == "1":
            return "fonalex45/toolkit:latest"
        elif choice == "2":
            return "fonalex45/toolkit:dev"
        elif choice == "3" and most_recent:
            return f"fonalex45/toolkit:{most_recent}"
        else:
            custom_tag = Prompt.ask("Enter custom tag")
            if not custom_tag.startswith("fonalex45/toolkit:"):
                custom_tag = f"fonalex45/toolkit:{custom_tag}"
            return custom_tag

    def run_cli(self):
        """Run CLI mode"""
        parser = argparse.ArgumentParser(
            description='Security Research Container Manager',
            formatter_class=argparse.RawDescriptionHelpFormatter
        )

        subparsers = parser.add_subparsers(dest='command', help='Commands')

        # Start command
        start_parser = subparsers.add_parser(
            'start', help='Start a new container')
        start_parser.add_argument('project', help='Project name')
        start_parser.add_argument('--image', help='Docker image to use')
        start_parser.add_argument(
            '--select-tag', action='store_true', help='Interactively select Docker tag')

        # Enter command
        enter_parser = subparsers.add_parser('enter', help='Enter a container')
        enter_parser.add_argument('project', help='Project name')

        # Stop command
        stop_parser = subparsers.add_parser('stop', help='Stop a container')
        stop_parser.add_argument('project', help='Project name')

        # Destroy command
        destroy_parser = subparsers.add_parser(
            'destroy', help='Destroy a container')
        destroy_parser.add_argument('project', help='Project name')
        destroy_parser.add_argument(
            '--force', action='store_true', help='Force destroy without confirmation')

        # Backup command
        backup_parser = subparsers.add_parser(
            'backup', help='Backup project data')
        backup_parser.add_argument('project', help='Project name')
        backup_parser.add_argument(
            '--dir', default='./backups', help='Backup directory')

        # Pull command
        pull_parser = subparsers.add_parser('pull', help='Pull Docker image')
        pull_parser.add_argument('--image', help='Image to pull')
        pull_parser.add_argument(
            '--select-tag', action='store_true', help='Interactively select Docker tag')

        # List command
        subparsers.add_parser('list', help='List all containers')

        # Config command
        subparsers.add_parser('config', help='Manage configuration')

        args = parser.parse_args()

        if not args.command:
            self.show_banner()
            parser.print_help()
            return

        # Check Docker availability
        if args.command not in ['config', 'tui']:
            if not self.docker_manager.check_docker():
                console.print(
                    f"[{Colors.ERROR}]Docker is not available or not running[/]")
                sys.exit(1)

        # Execute commands
        if args.command == 'start':
            image = args.image
            if args.select_tag:
                image = self.select_docker_tag()
            self.docker_manager.start_container(args.project, image)

        elif args.command == 'enter':
            self.docker_manager.enter_container(args.project)

        elif args.command == 'stop':
            self.docker_manager.stop_container(args.project)

        elif args.command == 'destroy':
            self.docker_manager.destroy_container(args.project, args.force)

        elif args.command == 'backup':
            self.docker_manager.backup_project(args.project, args.dir)

        elif args.command == 'pull':
            image = args.image
            if args.select_tag or not image:
                image = self.select_docker_tag()
            self.docker_manager.pull_image(image)

        elif args.command == 'list':
            containers = self.docker_manager.list_containers()
            if containers:
                table = Table(title="Toolkit Containers", show_header=True)
                table.add_column("Name", style="cyan")
                table.add_column("Status", style="yellow")
                table.add_column("Image", style="green")

                for container in containers:
                    table.add_row(
                        container['name'],
                        container['status'],
                        container['image']
                    )
                console.print(table)
            else:
                console.print(f"[{Colors.INFO}]No toolkit containers found[/]")

        elif args.command == 'config':
            self.manage_config()

    def manage_config(self):
        """Interactive configuration management"""
        console.print(f"\n[{Colors.HEADER}]Configuration Management[/]")

        table = Table(title="Current Configuration", show_header=True)
        table.add_column("Setting", style="cyan")
        table.add_column("Value", style="green")

        for key, value in asdict(self.config).items():
            if key != 'config_file':
                table.add_row(key, str(value))

        console.print(table)

        if Confirm.ask("\nModify configuration?"):
            self.config.docker_image = Prompt.ask(
                "Docker image", default=self.config.docker_image)
            self.config.container_shell = Prompt.ask(
                "Container shell", default=self.config.container_shell)
            self.config.host_networking = Confirm.ask(
                "Enable host networking?", default=self.config.host_networking)
            self.config.enable_x11 = Confirm.ask(
                "Enable X11 forwarding?", default=self.config.enable_x11)
            self.config.enable_gpu = Confirm.ask(
                "Enable GPU support?", default=self.config.enable_gpu)

            caps = Prompt.ask("Custom capabilities (comma-separated)",
                              default=','.join(self.config.custom_caps))
            self.config.custom_caps = [cap.strip() for cap in caps.split(',')]

            self.config.work_dir_base = Prompt.ask(
                "Work directory base", default=self.config.work_dir_base)

            self.config.save()


def main():
    """Main entry point"""
    cli = CLI()

    # Check if running with no arguments
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
