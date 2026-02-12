# Google Cloud VPS Setup Documentation

This document outlines the steps taken to set up a VPS on Google Cloud Platform (GCP) for the Space P2P project.

## 1. Install Google Cloud CLI

To interact with Google Cloud resources from the command line, install the `gcloud` CLI using `snap`:

```bash
sudo snap install google-cloud-cli --classic
```

## 2. Initialize GCloud

Initialize the CLI and authenticate with your Google account:

```bash
gcloud init
```

Follow the prompts to:
1.  Login to your Google account.
2.  Select the project (`space-487202`).
3.  Set a default zone (e.g., `asia-southeast1-a`).

## 3. Create VPS Instance

Create a Compute Engine instance with the following specifications:
*   **Name**: `space-p2p-vps`
*   **Zone**: `asia-southeast1-a`
*   **Machine Type**: `e2-medium` (2 vCPUs, 4GB RAM)
*   **Image**: Ubuntu 22.04 LTS (`ubuntu-2204-lts`)
*   **Tags**: `http-server`, `https-server`, `space-p2p`

Command used:

```bash
gcloud compute instances create space-p2p-vps \
    --zone=asia-southeast1-a \
    --machine-type=e2-medium \
    --image-family=ubuntu-2204-lts \
    --image-project=ubuntu-os-cloud \
    --tags=http-server,https-server,space-p2p
```

## 4. Configure Firewall Rules

Open the necessary ports for the application and LiveKit server:
*   **TCP 8080**: Space P2P API
*   **TCP 7880**: LiveKit HTTP API
*   **TCP 7881**: LiveKit WebRTC TCP
*   **UDP 50000-60000**: LiveKit WebRTC UDP

Command used:

```bash
gcloud compute firewall-rules create space-p2p-ports \
    --allow tcp:8080,tcp:7880,tcp:7881,udp:50000-60000 \
    --target-tags=space-p2p \
    --description="Allow Space P2P and LiveKit ports"
```

## 5. Setting Up SSH Access

Instead of relying solely on `gcloud compute ssh`, we set up standard SSH keys for easier automation.

### 5.1 Generate SSH Key Pair

Create a new SSH key specifically for GCP:

```bash
ssh-keygen -t rsa -f ~/.ssh/gcp_key -C htnguyen -N ""
```

### 5.2 Add Public Key to GCP Project Metadata

Add the public key to the project-wide metadata so it applies to all instances (including `space-p2p-vps`).

1.  Format the key for metadata (prefix with username):
    ```bash
    cat ~/.ssh/gcp_key.pub | awk '{print "htnguyen:" $0}' > ~/.ssh/gcp_metadata
    ```

2.  Upload the metadata:
    ```bash
    gcloud compute project-info add-metadata --metadata-from-file ssh-keys=~/.ssh/gcp_metadata
    ```

### 5.3 Verify Connection

Test the SSH connection:

```bash
ssh -i ~/.ssh/gcp_key -o StrictHostKeyChecking=no htnguyen@34.143.219.209 "echo 'Connection successful'"
```

*(Note: Replace `34.143.219.209` with your actual VPS IP address)*

## 6. Access & Deployment

### Accessing the VPS

To SSH into the VPS:

```bash
ssh -i ~/.ssh/gcp_key htnguyen@<VPS_IP>
```

### Deploying the Application

Use the `deployment.sh` script. Since the script uses the default `ssh` command, you need to ensure your SSH agent is aware of the key:

1.  Start the SSH agent (if not already running):
    ```bash
    eval $(ssh-agent)
    ```

2.  Add the key to the agent:
    ```bash
    ssh-add ~/.ssh/gcp_key
    ```

3.  Run the deployment script:
    ```bash
    ./deployment.sh htnguyen <VPS_IP>
    ```
