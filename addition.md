CLAUD.md: Building scrt - A Docker Wrapper in Go
This document outlines the Continuous Learning, Adaptation, and Understanding Document (CLAUD) for building scrt, a Docker wrapper in Go.

1. Project Overview
scrt is a command-line utility designed to simplify and enhance interaction with the Docker environment. It provides an interactive terminal user interface (TUI) for managing containers, offering a more user-friendly experience than the standard Docker CLI for common operations. This tool is particularly useful for tasks related to container lifecycle management, file operations, and forensic analysis.

2. Features
The scrt docker wrapper will provide the following functionalities:

List and Select Containers: Display a list of all containers (running and stopped) and allow the user to select one to start and enter.

Start Container: Start a stopped container.

Stop Container: Stop a running container.

Exec into Container: Open an interactive shell session within a running container.

Copy Files: Copy files from a container to the local filesystem.

Backup Container: Create a tar archive of a container's filesystem. This is particularly useful for forensic analysis. 

Import Backup: Import a container from a tar archive. 

Interactive Image Pull: An interactive menu to pull Docker images with different tags (e.g., latest, dev, or a custom tag).

3. Libraries and Technologies
The following Go libraries will be utilized for building scrt:

Docker Go SDK (github.com/docker/docker/client): The official Go SDK for the Docker API. It will be used to interact with the Docker daemon for all container and image operations. 
 


Lipgloss (github.com/charmbracelet/lipgloss): A library for styling terminal output with colors, borders, and other visual enhancements. 
 

Tview (github.com/rivo/tview): A rich interactive widget library for terminal-based user interfaces. 
 
 It will be used to create the interactive menus and lists for container selection and image pulling.

4. Implementation Details
This section details the implementation approach for each feature.

4.1. Core Docker Client
First, a Docker client needs to be initialized. The client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation()) function from the Docker Go SDK will be used to create a client that connects to the Docker daemon using environment variables. 

4.2. List and Select Containers
TUI: A tview.List will be used to display the list of containers. 
 

Docker Interaction: The cli.ContainerList() function from the Docker Go SDK will be called to get a list of all containers. 
 
 The types.ContainerListOptions{All: true} option should be used to list both running and stopped containers.

Styling: lipgloss will be used to style the list, highlighting the selected container. 

4.3. Start a Container
Docker Interaction: Once a container is selected from the tview.List, its ID will be used with the cli.ContainerStart() function to start it. 

4.4. Stop a Container
Docker Interaction: Similarly, to stop a running container, its ID will be passed to the cli.ContainerStop() function. 

4.5. Exec into a Container
Docker Interaction: The cli.ContainerExecCreate() function will create an exec instance in the container, and cli.ContainerExecAttach() will attach the terminal's stdin, stdout, and stderr to the exec process, providing an interactive shell. 
 

4.6. Copy Files from a Container
Docker Interaction: The cli.CopyFromContainer() function allows copying files and directories from a container's filesystem to the host. 

4.7. Backup Container to a Tar File
Docker Interaction: The cli.ContainerExport() method will be used to get a tar stream of the container's filesystem. 
 
 This stream will then be written to a .tar file on the local disk.

4.8. Import Tar Backup
Docker Interaction: The cli.ImageImport() function can import a tarball as a new image. 
 The tarball can be a local file or a URL. The function takes an types.ImageImportSource struct specifying the source and an types.ImageImportOptions struct for options like a new repository and tag name. 

4.9. Interactive Image Pull Menu
TUI: A tview.DropDown or a tview.List will be used to present the user with a choice of image tags to pull (e.g., 'latest', 'dev'). An tview.InputField can be used for custom tags. 

Docker Interaction: The cli.ImagePull() function will be used to pull the selected image from a registry. The output of the pull can be streamed to the terminal to show the progress. 
 

5. Inspiration
The following project serves as an inspiration for the TUI aspect of scrt:

awesome-tuis: https://github.com/rothgar/awesome-tuis?tab=readme-ov-file#dockerlxck8s

6. Conclusion
By leveraging the power of the Docker Go SDK and the user-friendly TUI libraries lipgloss and tview, scrt will provide a streamlined and intuitive way to manage Docker containers. The combination of these tools will enable the creation of a robust and visually appealing CLI application for both novice and experienced Docker users.