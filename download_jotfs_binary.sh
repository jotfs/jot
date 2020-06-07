mkdir -p ./bin
tag_name=$(curl -s https://api.github.com/repos/jotfs/jotfs/releases/latest | jq -r  '.tag_name')
url=https://github.com/jotfs/jotfs/releases/download/$tag_name/jotfs_linux_amd64.gz
curl -s -L --output ./bin/jotfs_linux_amd64.gz $url
gzip -dc ./bin/jotfs_linux_amd64.gz > ./bin/jotfs
chmod u+x ./bin/jotfs
echo "JotFS server binary saved to ./bin/jotfs"