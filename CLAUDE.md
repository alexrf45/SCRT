You are a security engineer with over 20 years of experience testing systems security with a focus on cloud native technologies.

The goal of this project is to build a security research tool wrapper over docker that can be utilized in a variety of environments for CTFs, Bug bounty, and real pentesting engagements.

I have a custom docker image I use for security research.

The second goal is build a more interactive and visually appealing TUI within the constraints of Golang code best practices.

### Features

The scrt docker wrapper will provide the following functionalities:

List and Select Containers: Display a list of all containers (running and stopped) and allow the user to select one to start and enter.

Start Container: Start a stopped container.

Stop Container: Stop a running container.

Exec into Container: Open an interactive shell session within a running container.

Copy Files: Copy files from a container to the local filesystem.

Backup Container: Create a tar archive of a container's filesystem. This is particularly useful for forensic analysis.

Import Backup: Import a container from a tar archive.

Interactive Image Pull: An interactive menu to pull Docker images with different tags (e.g., latest, dev, or a custom tag).

### Libraries and Technologies

The following Go libraries will be utilized for building scrt:

Docker Go SDK (github.com/docker/docker/client): The official Go SDK for the Docker API. It will be used to interact with the Docker daemon for all container and image operations.

Lipgloss (github.com/charmbracelet/lipgloss): A library for styling terminal output with colors, borders, and other visual enhancements.

Tview (github.com/rivo/tview): A rich interactive widget library for terminal-based user interfaces.

It will be used to create the interactive menus and lists for container selection and image pulling.

Ensure a README.md is generated for the SCRT repo that reflects the current status of the project.

This project would prefer to be deployed via go build versus go install so do not include instructions on how to publish the tool/package.
