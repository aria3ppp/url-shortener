# url-shortener

`url-shortener` is a backend implementation for a url shortener service written in golang advantage Hexagonal architecture (ports & adaptors) and DDD (domain-driven design).

[![Tests](https://github.com/aria3ppp/url-shortener/actions/workflows/tests.yml/badge.svg)](https://github.com/aria3ppp/url-shortener/actions/workflows/tests.yml)
[![Coverage Status](https://coveralls.io/repos/github/aria3ppp/url-shortener/badge.svg?branch=master)](https://coveralls.io/github/aria3ppp/url-shortener?branch=master)

### To test running the server using docker compose:

```bash
cp .env.example .env && make server-testdeploy-up
```
Now server is up and running at port `8080` on your `localhost`.

### To deploy on kubernetes (minikube) using terraform:
```bash
cd terraform/
terraform init
terraform apply
```
The resulting deployment consists of a `postgresql-ha` helm chart primary/reader instances and a load-balancing `url-shortener` service.
After which `url-shortener` srevice is available at `$(minikube ip):30000`