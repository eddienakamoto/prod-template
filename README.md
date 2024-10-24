# Prod Service Template
Prod Service Template is a sample project that shows the steps to get a production application deployed
with support for the following key infrastructure components:
- CI/CD of application code and supporting infrastructure config files using Github Actions
- Systemd support for application code
- Prometheus metrics for application code
- Automatic TLS with Caddy
- Log forwarding with Promtail from application code
- Centralized metric collection with Prometheus
- Centralized log collection with Loki
- Centralized dashboard with Grafana
- Sensitive data managed with Github Secrets

**This template assumes you deployed a server and configured your DNS to point to the appropriate IP**

## Github Secrets
Sensitive data needed for deployment will be managed using Github Secrets. Some of these variables will be injected into
the environment on your production server during the Github Actions workflow. Below is a table of the secrets this deployment
depends on.

| Secret          | Description                                                           |
| :-------------- | :-------------------------------------------------------------------- |
| SSH_PRIVATE_KEY | The private key used to SSH into the server during the CI/CD pipeline |
| SSH_USER        | The user to SSH into the production server as                         |
| SSH_SERVER      | The IP address of your production server                              |
| DOMAIN          | The url the service will be serving for                               |
| HOST            | The address of the service on the production server                   |
| PORT            | The port the service is being served on                               |
| PROMTAIL_PORT   | The port promtail serves metrics on the production server             |
| LOKI_SERVER_URL | The url of the loki server that promtail will forward logs to         |

Any environment variables the application will depend on can be added to your secrets and exported to the `.env` file during
the `Create and export secrets to environment file` step of the `deploy-go-app` step. For example, to export a secret called
`MY_SECRET`, the following line will use the `MY_SECRET` secret and export it to the `.env` file as `MY_SECRET`:
```bash
echo "MY_SECRET=${{ secrets.MY_SECRET }}" >> /etc/prod-template/.env
```

## Caddy
To install caddy execute the following commands.
```bash
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https curl
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt update
sudo apt install caddy
```

Verify caddy was installed.
```bash
caddy --version
```

Verify it is running with `systemd`.
```bash
sudo systemctl status caddy
```

## Promtail
To install promtail execute the following commands.
```bash
mkdir -p /etc/apt/keyrings/
wget -q -O - https://apt.grafana.com/gpg.key | gpg --dearmor > /etc/apt/keyrings/grafana.gpg
echo "deb [signed-by=/etc/apt/keyrings/grafana.gpg] https://apt.grafana.com stable main" | tee /etc/apt/sources.list.d/grafana.list
apt-get update
apt-get install loki promtail
```

Verify promtail was installed.
```bash
promtail --version
```

Verify it is running with `systemd`.
```bash
sudo systemctl status promtail
```