import type { LeaseConfig } from "./config";

export function cloudInit(config: LeaseConfig): string {
  return `#cloud-config
package_update: true
package_upgrade: false
users:
  - name: ${config.sshUser}
    groups: sudo
    shell: /bin/bash
    sudo: ['ALL=(ALL) NOPASSWD:ALL']
    ssh_authorized_keys:
      - ${config.sshPublicKey}
packages:
  - openssh-server
  - ca-certificates
  - curl
  - git
  - rsync
  - build-essential
  - docker.io
  - jq
write_files:
  - path: /etc/ssh/sshd_config.d/99-crabbox-port.conf
    permissions: '0644'
    content: |
      Port 22
      Port ${config.sshPort}
      PasswordAuthentication no
  - path: /usr/local/bin/crabbox-ready
    permissions: '0755'
    content: |
      #!/usr/bin/env bash
      set -euo pipefail
      node --version
      pnpm --version
      git --version
      rsync --version >/dev/null
      docker --version
runcmd:
  - mkdir -p ${config.workRoot} /var/cache/crabbox/pnpm /var/cache/crabbox/npm
  - chown -R ${config.sshUser}:${config.sshUser} ${config.workRoot} /var/cache/crabbox
  - systemctl enable --now ssh
  - systemctl restart ssh
  - systemctl enable --now docker
  - usermod -aG docker ${config.sshUser}
  - curl -fsSL https://deb.nodesource.com/setup_22.x | bash -
  - apt-get install -y nodejs
  - corepack enable
  - corepack prepare pnpm@10.33.2 --activate
  - sudo -u ${config.sshUser} bash -lc 'pnpm config set store-dir /var/cache/crabbox/pnpm'
  - crabbox-ready
`;
}
