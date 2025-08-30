#!/usr/bin/env python3

import os
import sys
import json
import time
import shutil
import tarfile
import argparse
import subprocess
from pathlib import Path
from datetime import datetime
from typing import Optional, List, Dict, Tuple, Any
from dataclasses import dataclass, field, asdict

try:
    import rich
    from rich.console import Console
    from rich.table import Table
    from rich.panel import Panel
    from rich.prompt import Prompt, Confirm
    from rich.progress import Progress, SpinnerColumn, TextColumn, BarColumn, TimeRemainingColumn
    from rich.text import Text
    from rich.tree import Tree
    from rich.align import Align
    from rich.rule import Rule
    from rich import box
    from rich.style import Style
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

console = Console(
    color_system="auto",
    force_terminal=True,
    highlight=True,
    soft_wrap=True
)


class Theme:
    ERROR = Style(color="red", bold=True)
    SUCCESS = Style(color="green", bold=True)
    WARNING = Style(color="yellow", bold=True)
    INFO = Style(color="cyan")
    HEADER = Style(color="bright_cyan", bold=True)
    ACCENT = Style(color="magenta", bold=True)
    RUNNING = Style(color="green", bold=True)
    STOPPED = Style(color="yellow")

    ICONS = {
        'success': '✓',
        'error': '✗',
        'warning': '⚠',
        'info': 'ℹ',
        'docker': '🐳',
        'container': '📦',
        'folder': '📁',
        'config': '⚙',
        'backup': '💾',
        'network': '🌐',
        'security': '🔒',
        'running': '▶',
        'stopped': '⏸',
    }


@dataclass
class Config:
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
        if self.config_file.exists():
            try:
                with open(self.config_file, 'r') as f:
                    data = json.load(f)
                    for key, value in data.items():
                        if hasattr(self, key):
                            setattr(self, key, value)
                console.print(
                    Panel(
                        f"{Theme.ICONS['success']} Configuration loaded from [path]{
                            self.config_file}[/path]",
                        style=Theme.SUCCESS,
                        border_style="green"
                    )
                )
            except Exception as e:
                console.print(
                    Panel(
                        f"{Theme.ICONS['warning']} Failed to load config: {e}",
                        style=Theme.WARNING,
                        border_style="yellow"
                    )
                )

    def save(self):
        try:
            with open(self.config_file, 'w') as f:
                json.dump(asdict(self), f, indent=2, default=str)
            console.print(
                Panel(
                    f"{Theme.ICONS['success']} Configuration saved to [path]{
                        self.config_file}[/path]",
                    style=Theme.SUCCESS,
                    border_style="green"
                )
            )
        except Exception as e:
            console.print(
                Panel(
                    f"{Theme.ICONS['error']} Failed to save config: {e}",
                    style=Theme.ERROR,
                    border_style="red"
                )
            )


class DockerManager:
    def __init__(self, config: Config):
        self.config = config

    def check_docker(self) -> bool:
        try:
            result = subprocess.run(['docker', 'info'],
                                    capture_output=True,
                                    text=True,
                                    timeout=5)
            return result.returncode == 0
        except (subprocess.SubprocessError, FileNotFoundError):
            return False

    def container_exists(self, project: str) -> bool:
        try:
            result = subprocess.run(['docker', 'container', 'inspect', project],
                                    capture_output=True,
                                    text=True)
            return result.returncode == 0
        except subprocess.SubprocessError:
            return False

    def container_running(self, project: str) -> bool:
        try:
            result = subprocess.run(['docker', 'container', 'inspect', '-f',
                                     '{{.State.Running}}', project],
                                    capture_output=True,
                                    text=True)
            return result.stdout.strip() == 'true'
        except subprocess.SubprocessError:
            return False

    def get_container_info(self, project: str) -> Optional[Dict[str, Any]]:
        try:
            result = subprocess.run(['docker', 'container', 'inspect', project],
                                    capture_output=True,
                                    text=True)
            if result.returncode == 0:
                data = json.loads(result.stdout)[0]
                return {
                    'id': data['Id'][:12],
                    'created': data['Created'],
                    'status': data['State']['Status'],
                    'running': data['State']['Running'],
                    'image': data['Config']['Image'],
                    'mounts': len(data.get('Mounts', [])),
                    'networks': list(data.get('NetworkSettings', {}).get('Networks', {}).keys())
                }
        except (subprocess.SubprocessError, json.JSONDecodeError, IndexError):
            pass
        return None

    def get_available_tags(self) -> Tuple[List[str], Optional[str]]:
        with console.status("[cyan]Fetching available tags from Docker Hub...", spinner="dots"):
            try:
                url = "https://hub.docker.com/v2/repositories/fonalex45/toolkit/tags"
                response = requests.get(url, timeout=5)

                if response.status_code == 200:
                    data = response.json()
                    tags = [tag['name'] for tag in data.get('results', [])]

                    priority_tags = []
                    version_tags = []

                    for tag in tags:
                        if tag in ['latest', 'dev']:
                            priority_tags.append(tag)
                        elif tag.startswith('v'):
                            version_tags.append(tag)

                    version_tags.sort(
                        key=lambda x: x.lstrip('v'), reverse=True)
                    most_recent_version = version_tags[0] if version_tags else None

                    return priority_tags + version_tags, most_recent_version
                else:
                    return ['latest', 'dev'], None

            except Exception as e:
                console.print(f"[dim]Could not fetch tags: {e}[/dim]")
                return ['latest', 'dev'], None

    def pull_image(self, image: str) -> bool:
        try:
            with Progress(
                SpinnerColumn(spinner_name="dots", style="cyan"),
                TextColumn("[progress.description]{task.description}"),
                BarColumn(bar_width=40),
                TextColumn("[progress.percentage]{task.percentage:>3.0f}%"),
                TimeRemainingColumn(),
                console=console,
                transient=True
            ) as progress:
                task = progress.add_task(f"Pulling {image}", total=100)

                process = subprocess.Popen(
                    ['docker', 'pull', image],
                    stdout=subprocess.PIPE,
                    stderr=subprocess.STDOUT,
                    text=True
                )

                for line in iter(process.stdout.readline, ''):
                    if line:
                        progress.advance(task, 10)

                process.wait()

                if process.returncode == 0:
                    console.print(
                        Panel(
                            f"{Theme.ICONS['success']} Successfully pulled [bold cyan]{
                                image}[/bold cyan]",
                            style=Theme.SUCCESS,
                            border_style="green",
                            title="[green]Pull Complete[/green]",
                            title_align="left"
                        )
                    )
                    return True
                else:
                    console.print(
                        Panel(
                            f"{Theme.ICONS['error']} Failed to pull image",
                            style=Theme.ERROR,
                            border_style="red"
                        )
                    )
                    return False

        except subprocess.SubprocessError as e:
            console.print(
                Panel(
                    f"{Theme.ICONS['error']} Error pulling image: {e}",
                    style=Theme.ERROR,
                    border_style="red"
                )
            )
            return False

    def start_container(self, project: str, image: Optional[str] = None) -> bool:
        if not self.validate_project_name(project):
            return False

        if self.container_exists(project):
            info = self.get_container_info(project)
            status_icon = Theme.ICONS['running'] if info and info['running'] else Theme.ICONS['stopped']
            status_text = "running" if info and info['running'] else "stopped"

            console.print(
                Panel(
                    f"{Theme.ICONS['warning']} Container '[bold]{
                        project}[/bold]' already exists\n"
                    f"Status: {status_icon} {status_text}\n"
                    f"Use 'enter' command to access it",
                    style=Theme.WARNING,
                    border_style="yellow",
                    title="[yellow]Container Exists[/yellow]",
                    title_align="left"
                )
            )
            return False

        project_dir = Path(self.config.work_dir_base) / project
        dirs = ['recon', 'www', 'exploit', 'pivot',
                'privesc', 'report', '.gr3ysh3ll-logs']

        tree = Tree(f"{Theme.ICONS['folder']} [bold cyan]{
                    project}[/bold cyan]")
        for dir_name in dirs:
            dir_path = project_dir / dir_name
            dir_path.mkdir(parents=True, exist_ok=True)
            icon = Theme.ICONS['security'] if dir_name.startswith(
                '.') else Theme.ICONS['folder']
            tree.add(f"{icon} {dir_name}")

        console.print(
            Panel(
                tree,
                title="[cyan]Project Structure Created[/cyan]",
                border_style="cyan",
                padding=(1, 2)
            )
        )

        image = image or self.config.docker_image
        cmd = ['docker', 'run', '--name', project, '-it']

        config_items = []

        if self.config.host_networking:
            cmd.append('--net=host')
            config_items.append(f"{Theme.ICONS['network']} Host Networking")

        for cap in self.config.custom_caps:
            cmd.append(f'--cap-add={cap}')
        if self.config.custom_caps:
            config_items.append(f"{Theme.ICONS['security']} Capabilities: {
                                ', '.join(self.config.custom_caps)}")

        if self.config.enable_gpu and Path('/dev/dri').exists():
            cmd.append('--device=/dev/dri:/dev/dri')
            config_items.append(f"🎮 GPU Support")

        if self.config.enable_x11 and os.environ.get('DISPLAY'):
            cmd.extend(['-e', f'DISPLAY={os.environ["DISPLAY"]}'])
            cmd.extend(['-v', '/tmp/.X11-unix:/tmp/.X11-unix'])
            xauth = Path.home() / '.Xauthority'
            if xauth.exists():
                cmd.extend(['-v', f'{xauth}:{xauth}'])
            config_items.append(f"🖥️ X11 Forwarding")

        cmd.extend(['-e', f'TARGET={project}'])
        cmd.extend(['-e', f'TZ={os.environ.get("TZ", "UTC")}'])

        cmd.extend(['-v', f'{project_dir}/.gr3ysh3ll-logs:/root/.logs:rw'])
        cmd.extend(['-v', f'{project_dir}:/{project}'])

        for mount in self.config.extra_mounts:
            cmd.extend(['-v', mount])

        cmd.extend(['-w', f'/{project}'])
        cmd.extend(['--entrypoint', self.config.container_shell])
        cmd.append(image)

        config_panel = Panel(
            "\n".join(config_items),
            title=f"[cyan]Starting Container: {project}[/cyan]",
            subtitle=f"[dim]Image: {image}[/dim]",
            border_style="cyan",
            padding=(1, 2)
        )
        console.print(config_panel)

        try:
            subprocess.run(cmd)
            return True
        except subprocess.SubprocessError as e:
            console.print(
                Panel(
                    f"{Theme.ICONS['error']} Failed to start container: {e}",
                    style=Theme.ERROR,
                    border_style="red"
                )
            )
            return False

    def enter_container(self, project: str) -> bool:
        if not self.container_exists(project):
            console.print(
                Panel(
                    f"{Theme.ICONS['error']} Container '[bold]{
                        project}[/bold]' does not exist",
                    style=Theme.ERROR,
                    border_style="red",
                    title="[red]Container Not Found[/red]"
                )
            )
            return False

        info = self.get_container_info(project)

        if not self.container_running(project):
            with console.status(f"[cyan]Starting container '{project}'...", spinner="dots"):
                subprocess.run(['docker', 'container', 'start',
                               project], capture_output=True)
                time.sleep(1)

        if info:
            info_table = Table(show_header=False,
                               box=box.SIMPLE, padding=(0, 1))
            info_table.add_column(style="cyan")
            info_table.add_column()

            info_table.add_row("Container ID", info['id'])
            info_table.add_row("Image", info['image'])
            info_table.add_row("Status", f"{Theme.ICONS['running']} Running" if info['running'] else f"{
                               Theme.ICONS['stopped']} Stopped")
            info_table.add_row("Networks", ", ".join(info['networks']))
            info_table.add_row("Mounts", str(info['mounts']))

            console.print(
                Panel(
                    info_table,
                    title=f"[cyan]Entering Container: {project}[/cyan]",
                    border_style="cyan",
                    padding=(1, 2)
                )
            )

        console.print(
            f"\n{Theme.ICONS['container']} Attaching to container...\n", style="cyan")

        try:
            subprocess.run(['docker', 'exec', '-it', project,
                           self.config.container_shell])
            return True
        except subprocess.SubprocessError as e:
            console.print(
                Panel(
                    f"{Theme.ICONS['error']} Failed to enter container: {e}",
                    style=Theme.ERROR,
                    border_style="red"
                )
            )
            return False

    def stop_container(self, project: str) -> bool:
        if not self.container_exists(project):
            console.print(
                Panel(
                    f"{Theme.ICONS['warning']} Container '[bold]{
                        project}[/bold]' does not exist",
                    style=Theme.WARNING,
                    border_style="yellow"
                )
            )
            return False

        if self.container_running(project):
            with console.status(f"[yellow]Stopping container '{project}'...", spinner="dots"):
                try:
                    subprocess.run(
                        ['docker', 'container', 'stop', project], capture_output=True)
                    time.sleep(1)

                    console.print(
                        Panel(
                            f"{Theme.ICONS['stopped']} Container '[bold]{
                                project}[/bold]' stopped successfully",
                            style=Theme.SUCCESS,
                            border_style="green",
                            title="[green]Container Stopped[/green]"
                        )
                    )
                    return True
                except subprocess.SubprocessError as e:
                    console.print(
                        Panel(
                            f"{Theme.ICONS['error']
                               } Failed to stop container: {e}",
                            style=Theme.ERROR,
                            border_style="red"
                        )
                    )
                    return False
        else:
            console.print(
                Panel(
                    f"{Theme.ICONS['info']} Container '[bold]{
                        project}[/bold]' is already stopped",
                    style=Theme.INFO,
                    border_style="cyan"
                )
            )
            return True

    def destroy_container(self, project: str, force: bool = False) -> bool:
        project_dir = Path(self.config.work_dir_base) / project

        if not self.container_exists(project):
            if project_dir.exists():
                if not force:
                    console.print(
                        Panel(
                            f"{Theme.ICONS['warning']
                               } Container doesn't exist but project directory found:\n"
                            f"[path]{project_dir}[/path]",
                            style=Theme.WARNING,
                            border_style="yellow",
                            title="[yellow]Orphaned Directory[/yellow]"
                        )
                    )
                    if not Confirm.ask(f"Remove project directory?", default=False):
                        return False

                shutil.rmtree(project_dir)
                console.print(
                    Panel(
                        f"{Theme.ICONS['success']} Project directory removed",
                        style=Theme.SUCCESS,
                        border_style="green"
                    )
                )
            else:
                console.print(
                    Panel(
                        f"{Theme.ICONS['info']} Container '[bold]{
                            project}[/bold]' does not exist",
                        style=Theme.INFO,
                        border_style="cyan"
                    )
                )
            return True

        info = self.get_container_info(project)

        if not force and info:
            destruction_tree = Tree(f"[red]⚠️  Destruction Summary[/red]")

            container_branch = destruction_tree.add(
                f"{Theme.ICONS['container']} Container")
            container_branch.add(f"Name: [bold]{project}[/bold]")
            container_branch.add(f"ID: {info['id']}")
            container_branch.add(f"Image: {info['image']}")

            if project_dir.exists():
                dir_branch = destruction_tree.add(
                    f"{Theme.ICONS['folder']} Project Directory")
                dir_branch.add(f"Path: [path]{project_dir}[/path]")

                total_size = sum(
                    f.stat().st_size for f in project_dir.rglob('*') if f.is_file())
                dir_branch.add(f"Size: {total_size / 1024 / 1024:.2f} MB")

            console.print(
                Panel(
                    destruction_tree,
                    title="[red]DESTRUCTIVE ACTION[/red]",
                    border_style="red",
                    padding=(1, 2)
                )
            )

            if not Confirm.ask(f"\n[bold red]Permanently destroy container '{project}' and all its data?[/bold red]", default=False):
                console.print("[cyan]Operation cancelled[/cyan]")
                return False

        with console.status(f"[red]Destroying container '{project}'...", spinner="dots"):
            try:
                subprocess.run(['docker', 'container', 'rm',
                               '-f', project], capture_output=True)

                if project_dir.exists():
                    shutil.rmtree(project_dir)

                time.sleep(1)

                console.print(
                    Panel(
                        f"{Theme.ICONS['success']} Container '[bold]{
                            project}[/bold]' destroyed permanently",
                        style=Theme.SUCCESS,
                        border_style="green",
                        title="[green]Destruction Complete[/green]"
                    )
                )
                return True

            except (subprocess.SubprocessError, OSError) as e:
                console.print(
                    Panel(
                        f"{Theme.ICONS['error']
                           } Failed to destroy container: {e}",
                        style=Theme.ERROR,
                        border_style="red"
                    )
                )
                return False

    def backup_project(self, project: str, backup_dir: str = "./backups") -> bool:
        project_path = Path(self.config.work_dir_base) / project

        if not project_path.exists():
            console.print(
                Panel(
                    f"{Theme.ICONS['error']} Project directory '[path]{
                        project_path}[/path]' does not exist",
                    style=Theme.ERROR,
                    border_style="red"
                )
            )
            return False

        backup_path = Path(backup_dir)
        backup_path.mkdir(parents=True, exist_ok=True)

        timestamp = datetime.now().strftime('%Y-%m-%d_%H-%M-%S')
        backup_file = backup_path / f"{timestamp}_{project}.tar.gz"

        total_size = sum(
            f.stat().st_size for f in project_path.rglob('*') if f.is_file())

        console.print(
            Panel(
                f"{Theme.ICONS['backup']} Creating backup of '[bold]{
                    project}[/bold]'\n"
                f"Source: [path]{project_path}[/path]\n"
                f"Size: {total_size / 1024 / 1024:.2f} MB",
                title="[cyan]Backup Started[/cyan]",
                border_style="cyan"
            )
        )

        try:
            with Progress(
                SpinnerColumn(spinner_name="dots", style="cyan"),
                TextColumn("[progress.description]{task.description}"),
                BarColumn(bar_width=40),
                TextColumn("[progress.percentage]{task.percentage:>3.0f}%"),
                console=console,
                transient=True
            ) as progress:
                task = progress.add_task("Compressing files...", total=100)

                with tarfile.open(backup_file, 'w:gz') as tar:
                    tar.add(project_path, arcname=project)
                    for i in range(10):
                        time.sleep(0.1)
                        progress.advance(task, 10)

            console.print(
                Panel(
                    f"{Theme.ICONS['success']} Backup created successfully\n"
                    f"File: [path]{backup_file}[/path]\n"
                    f"Size: {backup_file.stat().st_size / 1024 / 1024:.2f} MB",
                    style=Theme.SUCCESS,
                    border_style="green",
                    title="[green]Backup Complete[/green]"
                )
            )
            return True

        except Exception as e:
            console.print(
                Panel(
                    f"{Theme.ICONS['error']} Failed to create backup: {e}",
                    style=Theme.ERROR,
                    border_style="red"
                )
            )
            return False

    def list_containers(self) -> List[Dict[str, str]]:
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
                            'image': data.get('Image', ''),
                            'state': data.get('State', ''),
                            'created': data.get('CreatedAt', '')
                        })
                    except json.JSONDecodeError:
                        continue

            return containers

        except subprocess.SubprocessError:
            return []

    @staticmethod
    def validate_project_name(project: str) -> bool:
        import re
        if not re.match(r'^[a-zA-Z0-9_-]+$', project):
            console.print(
                Panel(
                    f"{Theme.ICONS['error']} Invalid project name\n"
                    "Use only alphanumeric characters, hyphens, and underscores",
                    style=Theme.ERROR,
                    border_style="red"
                )
            )
            return False
        return True


class CLI:
    def __init__(self):
        self.config = Config()
        self.config.load()
        self.docker_manager = DockerManager(self.config)

    def show_banner(self):
        banner_text = """[bold cyan]
   [bold cyan]
             ________________________________________________
            //                                               \\
           |    _________________________________________     |
           |   |                                         |    |
           |   |  admin@toolkit $ _                      |    |
           |   |  Security Research Container Manager    |    |
           |   |                                         |    |
           |   |_________________________________________|    |
           |                                                  |
            \\_________________________________________________//
                   \\___________________________________//
[/]
        """

        console.print(Align.center(banner_text))
        console.print(Align.center(
            "[dim]Type 'toolkit-cli --help' for usage[/dim]\n"))

    def select_docker_tag(self) -> str:
        console.print(
            Rule(f"{Theme.ICONS['docker']} Docker Tag Selection", style="cyan"))

        with console.status("[cyan]Fetching available tags...", spinner="dots"):
            tags, most_recent = self.docker_manager.get_available_tags()
            time.sleep(0.5)

        option_panel = Panel(
            f"[bold green]1.[/bold green] latest - [dim]Latest stable release[/dim]\n"
            f"[bold yellow]2.[/bold yellow] dev - [dim]Development version[/dim]\n" +
            (f"[bold cyan]3.[/bold cyan] {most_recent} - [dim]Most recent version[/dim]\n" if most_recent else "") +
            f"[bold magenta]{
                4 if most_recent else 3}.[/bold magenta] custom - [dim]Enter custom tag[/dim]\n\n"
            f"[dim]View all tags: https://hub.docker.com/r/fonalex45/toolkit/tags[/dim]",
            title="[cyan]Available Options[/cyan]",
            border_style="cyan",
            padding=(1, 2)
        )

        console.print(option_panel)

        choices = ["1", "2", "3", "4"] if most_recent else ["1", "2", "3"]
        choice = Prompt.ask(
            f"\n{Theme.ICONS['docker']} Select an option",
            choices=choices,
            default="1"
        )

        if choice == "1":
            selected = "fonalex45/toolkit:latest"
        elif choice == "2":
            selected = "fonalex45/toolkit:dev"
        elif choice == "3" and most_recent:
            selected = f"fonalex45/toolkit:{most_recent}"
        else:
            custom_tag = Prompt.ask(
                f"{Theme.ICONS['docker']} Enter custom tag")
            if not custom_tag.startswith("fonalex45/toolkit:"):
                custom_tag = f"fonalex45/toolkit:{custom_tag}"
            selected = custom_tag

        console.print(
            Panel(
                f"{Theme.ICONS['success']} Selected: [bold cyan]{
                    selected}[/bold cyan]",
                style=Theme.SUCCESS,
                border_style="green"
            )
        )
        return selected

    def show_quick_help(self):
        help_content = Table(
            show_header=True, header_style="bold cyan", box=box.ROUNDED, padding=(0, 1))
        help_content.add_column("Command", style="bold green", width=20)
        help_content.add_column("Description", style="white")
        help_content.add_column("Example", style="dim")

        help_content.add_row(
            "start",
            "Create and start a new container",
            "toolkit-cli start myproject"
        )
        help_content.add_row(
            "enter",
            "Enter an existing container",
            "toolkit-cli enter myproject"
        )
        help_content.add_row(
            "stop",
            "Stop a running container",
            "toolkit-cli stop myproject"
        )
        help_content.add_row(
            "destroy",
            "Remove container and data",
            "toolkit-cli destroy myproject"
        )
        help_content.add_row(
            "backup",
            "Backup project data",
            "toolkit-cli backup myproject"
        )
        help_content.add_row(
            "list",
            "List all containers",
            "toolkit-cli list"
        )
        help_content.add_row(
            "status",
            "Show container status",
            "toolkit-cli status myproject"
        )
        help_content.add_row(
            "config",
            "Manage configuration",
            "toolkit-cli config"
        )

        console.print(
            Panel(
                help_content,
                title=f"{Theme.ICONS['info']} Quick Command Reference",
                border_style="cyan",
                padding=(1, 1)
            )
        )

    def show_container_list(self):
        containers = self.docker_manager.list_containers()

        if not containers:
            console.print(
                Panel(
                    f"{Theme.ICONS['info']} No toolkit containers found\n\n"
                    "Create your first container with:\n"
                    "  [bold cyan]toolkit-cli start <project-name>[/bold cyan]",
                    style=Theme.INFO,
                    border_style="cyan",
                    title="[cyan]No Containers[/cyan]"
                )
            )
            return

        table = Table(
            show_header=True,
            header_style="bold cyan",
            box=box.ROUNDED,
            title=f"{Theme.ICONS['container']} Toolkit Containers",
            title_style="bold cyan",
            padding=(0, 1)
        )

        table.add_column("Name", style="bold white", width=20)
        table.add_column("Status", justify="center", width=15)
        table.add_column("Image", style="dim", width=30)
        table.add_column("Created", style="dim", width=20)

        for container in containers:
            status = container['state'].lower()
            if 'running' in status:
                status_display = f"[green]{
                    Theme.ICONS['running']} Running[/green]"
            elif 'exited' in status:
                status_display = f"[yellow]{
                    Theme.ICONS['stopped']} Stopped[/yellow]"
            else:
                status_display = f"[red]{status.title()}[/red]"

            table.add_row(
                container['name'],
                status_display,
                container['image'],
                container.get('created', 'Unknown')
            )

        console.print(table)
        console.print(f"\n[dim]Total containers: {len(containers)}[/dim]")

    def show_status(self, project: Optional[str] = None):
        if project:
            if not self.docker_manager.container_exists(project):
                console.print(
                    Panel(
                        f"{Theme.ICONS['error']} Container '[bold]{
                            project}[/bold]' does not exist",
                        style=Theme.ERROR,
                        border_style="red"
                    )
                )
                return

            info = self.docker_manager.get_container_info(project)
            if info:
                status_table = Table(
                    show_header=False, box=box.SIMPLE, padding=(0, 1))
                status_table.add_column(style="cyan", width=20)
                status_table.add_column()

                status_icon = Theme.ICONS['running'] if info['running'] else Theme.ICONS['stopped']
                status_text = "Running" if info['running'] else "Stopped"
                status_color = "green" if info['running'] else "yellow"

                status_table.add_row("Container", f"[bold]{project}[/bold]")
                status_table.add_row("Status", f"[{status_color}]{status_icon} {
                                     status_text}[/{status_color}]")
                status_table.add_row("Container ID", info['id'])
                status_table.add_row("Image", info['image'])
                status_table.add_row("Created", info['created'][:19])
                status_table.add_row("Networks", ", ".join(info['networks']))
                status_table.add_row("Mount Points", str(info['mounts']))

                project_dir = Path(self.config.work_dir_base) / project
                if project_dir.exists():
                    total_size = sum(
                        f.stat().st_size for f in project_dir.rglob('*') if f.is_file())
                    status_table.add_row("Project Path", f"[path]{
                                         project_dir}[/path]")
                    status_table.add_row("Project Size", f"{
                                         total_size / 1024 / 1024:.2f} MB")

                console.print(
                    Panel(
                        status_table,
                        title=f"[cyan]{Theme.ICONS['info']
                                       } Container Status[/cyan]",
                        border_style="cyan",
                        padding=(1, 2)
                    )
                )
        else:
            containers = self.docker_manager.list_containers()
            running = sum(1 for c in containers if 'running' in c.get(
                'state', '').lower())
            stopped = len(containers) - running

            summary = f"""
{Theme.ICONS['docker']} Docker Status: [green]Active[/green]
{Theme.ICONS['container']} Total Containers: [bold]{len(containers)}[/bold]
  {Theme.ICONS['running']} Running: [green]{running}[/green]
  {Theme.ICONS['stopped']} Stopped: [yellow]{stopped}[/yellow]

{Theme.ICONS['config']} Configuration:
  Image: [cyan]{self.config.docker_image}[/cyan]
  Work Dir: [path]{self.config.work_dir_base}[/path]
            """

            console.print(
                Panel(
                    summary.strip(),
                    title=f"[cyan]{Theme.ICONS['info']} System Status[/cyan]",
                    border_style="cyan",
                    padding=(1, 2)
                )
            )

    def manage_config(self):
        console.print(
            Rule(f"{Theme.ICONS['config']} Configuration Management", style="cyan"))

        config_table = Table(
            show_header=True,
            header_style="bold cyan",
            box=box.ROUNDED,
            title="Current Configuration",
            padding=(0, 1)
        )
        config_table.add_column("Setting", style="cyan", width=25)
        config_table.add_column("Value", style="green")
        config_table.add_column("Description", style="dim")

        config_items = [
            ("docker_image", self.config.docker_image, "Docker image to use"),
            ("container_shell", self.config.container_shell,
             "Default shell in container"),
            ("host_networking", str(self.config.host_networking),
             "Use host network mode"),
            ("enable_x11", str(self.config.enable_x11), "Enable X11 forwarding"),
            ("enable_gpu", str(self.config.enable_gpu), "Enable GPU support"),
            ("custom_caps", ", ".join(self.config.custom_caps), "Linux capabilities"),
            ("work_dir_base", str(self.config.work_dir_base),
             "Base directory for projects"),
        ]

        for key, value, desc in config_items:
            config_table.add_row(key, value, desc)

        console.print(config_table)
        console.print(f"\n[dim]Config file: {self.config.config_file}[/dim]\n")

        if Confirm.ask(f"{Theme.ICONS['config']} Modify configuration?", default=False):
            console.print(Rule("Edit Configuration", style="yellow"))

            self.config.docker_image = Prompt.ask(
                f"{Theme.ICONS['docker']} Docker image",
                default=self.config.docker_image
            )

            self.config.container_shell = Prompt.ask(
                "🐚 Container shell",
                default=self.config.container_shell
            )

            self.config.host_networking = Confirm.ask(
                f"{Theme.ICONS['network']} Enable host networking?",
                default=self.config.host_networking
            )

            self.config.enable_x11 = Confirm.ask(
                "🖥️  Enable X11 forwarding?",
                default=self.config.enable_x11
            )

            self.config.enable_gpu = Confirm.ask(
                "🎮 Enable GPU support?",
                default=self.config.enable_gpu
            )

            caps = Prompt.ask(
                f"{Theme.ICONS['security']
                   } Custom capabilities (comma-separated)",
                default=','.join(self.config.custom_caps)
            )
            self.config.custom_caps = [cap.strip() for cap in caps.split(',')]

            self.config.work_dir_base = Prompt.ask(
                f"{Theme.ICONS['folder']} Work directory base",
                default=self.config.work_dir_base
            )

            self.config.save()

    def run_cli(self):
        parser = argparse.ArgumentParser(
            description='Security Research Container Manager',
            formatter_class=argparse.RawDescriptionHelpFormatter,
            epilog="For interactive mode, run without arguments"
        )

        subparsers = parser.add_subparsers(
            dest='command', help='Available commands')

        start_parser = subparsers.add_parser(
            'start', help='Start a new container')
        start_parser.add_argument('project', help='Project name')
        start_parser.add_argument('--image', help='Docker image to use')
        start_parser.add_argument(
            '--select-tag', action='store_true', help='Interactively select Docker tag')

        enter_parser = subparsers.add_parser('enter', help='Enter a container')
        enter_parser.add_argument('project', help='Project name')

        stop_parser = subparsers.add_parser('stop', help='Stop a container')
        stop_parser.add_argument('project', help='Project name')

        destroy_parser = subparsers.add_parser(
            'destroy', help='Destroy a container')
        destroy_parser.add_argument('project', help='Project name')
        destroy_parser.add_argument(
            '--force', action='store_true', help='Force destroy without confirmation')

        backup_parser = subparsers.add_parser(
            'backup', help='Backup project data')
        backup_parser.add_argument('project', help='Project name')
        backup_parser.add_argument(
            '--dir', default='./backups', help='Backup directory')

        pull_parser = subparsers.add_parser('pull', help='Pull Docker image')
        pull_parser.add_argument('--image', help='Image to pull')
        pull_parser.add_argument(
            '--select-tag', action='store_true', help='Interactively select Docker tag')

        subparsers.add_parser('list', help='List all containers')

        subparsers.add_parser('config', help='Manage configuration')

        status_parser = subparsers.add_parser(
            'status', help='Show container status')
        status_parser.add_argument(
            'project', nargs='?', help='Project name (optional)')

        args = parser.parse_args()

        if not args.command:
            self.show_banner()
            self.show_quick_help()
            return

        if args.command not in ['config']:
            if not self.docker_manager.check_docker():
                console.print(
                    Panel(
                        f"{Theme.ICONS['error']
                           } Docker is not available or not running\n\n"
                        "Please ensure Docker Desktop is running or Docker daemon is started:\n"
                        "  • macOS/Windows: Start Docker Desktop\n"
                        "  • Linux: sudo systemctl start docker",
                        style=Theme.ERROR,
                        border_style="red",
                        title="[red]Docker Not Found[/red]"
                    )
                )
                sys.exit(1)

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
            self.show_container_list()

        elif args.command == 'status':
            self.show_status(args.project)

        elif args.command == 'config':
            self.manage_config()


def main():
    try:
        cli = CLI()

        if len(sys.argv) == 1:
            cli.show_banner()

            if cli.docker_manager.check_docker():
                console.print(
                    f"[green]{Theme.ICONS['success']
                              } Docker is running[/green]\n",
                    justify="center"
                )
            else:
                console.print(
                    f"[red]{Theme.ICONS['error']
                            } Docker is not running[/red]\n",
                    justify="center"
                )

            cli.show_quick_help()
        else:
            cli.run_cli()

    except KeyboardInterrupt:
        console.print(
            f"\n[yellow]{Theme.ICONS['warning']
                         } Operation cancelled by user[/yellow]"
        )
        sys.exit(0)
    except Exception as e:
        console.print(
            Panel(
                f"{Theme.ICONS['error']} Unexpected error: {e}\n\n"
                "[dim]Please report this issue if it persists[/dim]",
                style=Theme.ERROR,
                border_style="red",
                title="[red]Fatal Error[/red]"
            )
        )
        if "--debug" in sys.argv:
            console.print_exception()
        sys.exit(1)


if __name__ == "__main__":
    main()

