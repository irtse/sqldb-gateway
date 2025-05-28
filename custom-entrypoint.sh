#!/bin/bash

# Copy the custom script into the Grafana public directory
cp /etc/grafana/custom.js /usr/share/grafana/public/custom.js

# Inject it into the main HTML file (only if not already injected)
if ! grep -q "custom.js" /usr/share/grafana/public/views/index.html; then
  sed -i '/<head>/a <script src="custom.js"></script>' /usr/share/grafana/public/views/index.html
fi

# Now run the original Grafana entrypoint
/run.sh