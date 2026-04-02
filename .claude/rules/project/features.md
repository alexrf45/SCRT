# SCRT Feature Requirements

These are the canonical features for SCRT. Do not remove or stub out any of these without explicit approval.

## Required Features

| Feature | Description |
|---------|-------------|
| List & Select Containers | Display all containers (running and stopped); allow user to select one to start and enter |
| Start Container | Start a stopped container |
| Stop Container | Stop a running container |
| Exec into Container | Open an interactive shell session within a running container |
| Copy Files | Copy files from a container to the local filesystem |
| Backup Container | Create a tar archive of a container's filesystem (useful for forensic analysis) |
| Import Backup | Import a container from a tar archive |
| Interactive Image Pull | Interactive menu to pull Docker images with selectable tags (latest, dev, custom) |

## Target Use Cases

SCRT is designed for:
- CTF (Capture the Flag) competitions
- Bug bounty engagements
- Real penetration testing engagements

Design decisions should favor these workflows. Assume the operator is technically proficient.
