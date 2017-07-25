#/bin/bash
mkdir -p ~/.paddle
cat > ~/.paddle/config << EOF
datacenters:
- name: datacenter1
  username: your-user-name
  password: your-secret
  endpoint: http://127.0.0.1:8080
current-datacenter: datacenter1
EOF
