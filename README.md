# Cloudflare Tunnel Ingress Controller

TLDR; This project simplifies exposing Kubernetes services to the internet easily and securely using Cloudflare Tunnel.

## Get Started

- You should have a Cloudflare account and a domain configured on Cloudflare.
- Create a Cloudflare API token with the following:
  - `Zone:Zone:Read`
  - `Zone:DNS:Edit`
  - `Account:Cloudflare Tunnel:Edit`
- Fetch the Account ID from the Cloudflare dashboard, follow the instructions [here](https://developers.cloudflare.com/fundamentals/get-started/basic-tasks/find-account-and-zone-ids/).
- Bootstrap a minikube cluster

```bash
minikube start
```

- Clone this project:

```bash
git clone https://github.com/STRRL/cloudflare-tunnel-ingress-controller && \
  cd cloudflare-tunnel-ingress-controller
```

- Install with Helm:

```bash
helm upgrade --install --wait \
  -n cloudflare-tunnel-ingress-controller --create-namespace \
  cloudflare-tunnel-ingress-controller \
  ./helm/cloudflare-tunnel-ingress-controller \
  --set=cloudflare.apiToken="<cloudflare-api-token>",cloudflare.accountId="<cloudflare-account-id>",cloudflare.tunnelName="<your-favorite-tunnel-name>" 
```

> if the tunnel does not exist, controller will create it for you.

- Then enable some awesome features in minikube, like kubernetes-dashboard:

```bash
minikube addons enable dashboard
minikube addons enable metrics-server
```

- Then expose the dashboard to the internet by creating an `Ingress`:

```bash
kubectl -n kubernetes-dashboard \
  create ingress dashboard-via-cf-tunnel \
  --rule="<your-favorite-domain>/*=kubernetes-dashboard:80"\
  --class cloudflare-tunnel
```

> for example, I would use `dash.strrl.cloud` as my favorite domain here.

- At last, access the dashboard via the domain you just created:

![dash.strrl.cloud](./static/dash.strrl.cloud.png)

- Done! Enjoy! 🎉
