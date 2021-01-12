Секрет нужно создавать следующей командой:
```shell
kubectl create secret docker-registry github-packages \
  --docker-server=docker.pkg.github.com \
  --docker-username=coursar \
  --docker-password=XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX \
  --docker-email=coursar@localhost
```

* `docker-username` - ваш логин на GitHub
* `docker-password` - [GitHub Personal Access Token](https://docs.github.com/en/free-pro-team@latest/github/authenticating-to-github/creating-a-personal-access-token) с правами на чтение пакетов (packages:read)
* `docker-email` - ваш email на GitHub

После этого уже применять конфигурацию с помощью `kubectl apply -f app.yml` (не забудьте перед этим `minikube start` и `minikube tunnel`).
