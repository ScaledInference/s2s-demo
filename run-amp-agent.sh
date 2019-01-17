#!/usr/bin/env bash
sudo docker pull scaledinference/ampagent:prod-latest
sudo docker run -d -t -i -e AMPAGENT_KEY=$1 -p 8100:8100 \
        --restart=on-failure:10 \
		--memory="1.5g" --memory-swap="1.5g" --sysctl net.core.somaxconn=1024 \
		--name ampagent scaledinference/ampagent:prod-latest