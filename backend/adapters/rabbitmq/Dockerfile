FROM rabbitmq:4.2-management

# Build-time deps to fetch latest plugin release metadata.
RUN apt-get update && apt-get install -y --no-install-recommends bash ca-certificates curl jq \
  && rm -rf /var/lib/apt/lists/*

# Download the latest delayed-message-exchange plugin .ez and enable it offline.
# The plugin is not bundled with the official image.
RUN bash -lc 'set -euo pipefail; \
  asset_url="$(curl -fsSL https://api.github.com/repos/rabbitmq/rabbitmq-delayed-message-exchange/releases/latest \
    | jq -r '"'"'.assets[] | select(.name | endswith(".ez")) | .browser_download_url'"'"' \
    | head -n 1)"; \
  test -n "$asset_url"; \
  curl -fsSL "$asset_url" -o "$RABBITMQ_HOME/plugins/rabbitmq_delayed_message_exchange.ez"; \
  rabbitmq-plugins enable --offline rabbitmq_delayed_message_exchange'

