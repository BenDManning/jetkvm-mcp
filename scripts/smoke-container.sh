#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 6 ]]; then
  echo "usage: smoke-container.sh IMAGE PLATFORM VERSION SOURCE REVISION CREATED" >&2
  exit 2
fi

image=$1
platform=$2
version=$3
source=$4
revision=$5
created=$6
architecture=${platform#linux/}
repository_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
config_path=${repository_root}/testdata/container/config.yaml
temporary_directory=$(mktemp -d)
container_id=

cleanup() {
  if [[ -n ${container_id} ]]; then
    docker logs "${container_id}" >&2 || true
    docker rm --force "${container_id}" >/dev/null 2>&1 || true
  fi
  rm -r -- "${temporary_directory}"
}
trap cleanup EXIT

inspect() {
  docker image inspect --format "$1" "${image}"
}

label() {
  inspect "{{ index .Config.Labels \"$1\" }}"
}

[[ $(inspect '{{.Architecture}}') == "${architecture}" ]]
[[ $(inspect '{{.Config.User}}') == "10001:10001" ]]
[[ $(inspect '{{json .Config.Entrypoint}}') == '["/usr/local/bin/jetkvm-mcp"]' ]]
[[ $(label org.opencontainers.image.source) == "${source}" ]]
[[ $(label org.opencontainers.image.revision) == "${revision}" ]]
[[ $(label org.opencontainers.image.version) == "${version}" ]]
[[ $(label org.opencontainers.image.licenses) == "MIT" ]]
[[ $(label org.opencontainers.image.created) == "${created}" ]]

if inspect '{{range .Config.Env}}{{println .}}{{end}}' | grep -Eiq '(^|_)(password|token|secret|credential)='; then
  echo "container image embeds a credential-shaped environment variable" >&2
  exit 1
fi

[[ $(docker run --rm "${image}" --version) == "jetkvm-mcp ${version}" ]]
[[ $(docker run --rm --entrypoint id "${image}" -u) == "10001" ]]
[[ $(docker run --rm --entrypoint id "${image}" -g) == "10001" ]]

docker run --rm --entrypoint sh "${image}" -euc '
  test ! -e /config.yaml
  test -s /usr/share/jetkvm-mcp/container-packages.txt
  test -s /usr/share/jetkvm-mcp/ffmpeg-version.txt
  grep -E "^ca-certificates(:[^[:space:]]+)?[[:space:]]" /usr/share/jetkvm-mcp/container-packages.txt
  grep -E "^ffmpeg(:[^[:space:]]+)?[[:space:]]" /usr/share/jetkvm-mcp/container-packages.txt
  cat /usr/share/jetkvm-mcp/container-packages.txt
  sed -n "1p" /usr/share/jetkvm-mcp/ffmpeg-version.txt
'

docker run --rm \
  --network none \
  --read-only \
  --tmpfs /tmp:rw,noexec,nosuid,size=64m,mode=1777 \
  --cap-drop ALL \
  --security-opt no-new-privileges \
  -e JETKVM_LAB_PASSWORD=ci-fixture \
  -e JETKVM_MCP_HTTP_TOKEN=ci-offline-token \
  -v "${config_path}:/config.yaml:ro" \
  "${image}" config validate --config /config.yaml

docker run --rm \
  --network none \
  --read-only \
  --tmpfs /tmp:rw,noexec,nosuid,size=64m,mode=1777 \
  --cap-drop ALL \
  --security-opt no-new-privileges \
  --entrypoint sh \
  "${image}" -euc '
    ffmpeg -nostdin -hide_banner -loglevel error \
      -f lavfi -i color=c=blue:s=64x48:d=0.1 -frames:v 1 \
      -c:v libx264 -preset ultrafast -tune zerolatency -f h264 /tmp/frame.h264
    ffmpeg -nostdin -hide_banner -loglevel error \
      -f h264 -i /tmp/frame.h264 -frames:v 1 -pix_fmt rgb24 -c:v png /tmp/frame.png
    test "$(ffprobe -v error -select_streams v:0 -show_entries stream=codec_name -of default=noprint_wrappers=1:nokey=1 /tmp/frame.png)" = png
    test "$(ffprobe -v error -select_streams v:0 -show_entries stream=width,height -of csv=p=0 /tmp/frame.png)" = 64,48
  '

container_id=$(docker run --detach \
  --read-only \
  --tmpfs /tmp:rw,noexec,nosuid,size=64m,mode=1777 \
  --cap-drop ALL \
  --security-opt no-new-privileges \
  --publish 127.0.0.1::8080 \
  -e JETKVM_LAB_PASSWORD=ci-fixture \
  -e JETKVM_MCP_HTTP_TOKEN=ci-health-token \
  -v "${config_path}:/config.yaml:ro" \
  "${image}" --config /config.yaml --http 0.0.0.0:8080)
endpoint=$(docker port "${container_id}" 8080/tcp | head -n 1)

healthy=false
for _ in {1..30}; do
  if curl --silent --fail \
    --dump-header "${temporary_directory}/headers" \
    --output "${temporary_directory}/body" \
    "http://${endpoint}/healthz"; then
    healthy=true
    break
  fi
  sleep 1
done
if [[ ${healthy} != true ]]; then
  echo "container health endpoint did not become available" >&2
  exit 1
fi
grep -Eiq '^content-type: text/plain; charset=utf-8\r?$' "${temporary_directory}/headers"
[[ $(<"${temporary_directory}/body") == "ok" ]]
