---
title: "Модуль user-authz: примеры использования"
---

## Создание пользователя

В Kubernetes есть две категории пользователей:
* ServiceAccount-ы, учёт которых ведёт сам Kubernetes через API.
* Остальные пользователи, чей учёт ведёт не сам Kubernetes, а некоторый внешний софт, который настраивает администратор кластера – существует множество механизмов аутентификации и, соответственно, множество способов заводить пользователей. В настоящий момент поддерживается два способа аутентификации:
    * Через модуль [user-authn](/modules/150-user-authn/).
    * С помощью сертификатов.

При выпуске сертификата для аутентификации, нужно  указать в нем `CN=<имя>`, необходимое количество `O=<группа>` и подписать его с помощью корневого CA кластера. Именно этим механизмом вы аутентифицируетесь в кластере, когда например используете kubectl на бастионе.

### Создаём пользователя с помощью клиентского сертификата

* Достаём корневой сертификат кластера (ca.crt и ca.key).
* Генерируем ключ пользователя:
```shell
openssl genrsa -out myuser.key 2048
```
* Создаём CSR, где указываем, что нам требуется пользователь `myuser`, который состоит в группах `mygroup1` и `mygroup2`:
```shell
openssl req -new -key myuser.key -out myuser.csr -subj "/CN=myuser/O=mygroup1/O=mygroup2"
```
* Подписываем CSR корневым сертификатом кластера:
```shell
openssl x509 -req -in myuser.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out myuser.crt -days 10000
```
* Теперь полученный сертификат можно указывать в конфиг-файле:
```shell
cat << EOF
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: $(cat ca.crt | base64 -w0)
    server: https://<хост кластера>:6443
  name: kubernetes
contexts:
- context:
    cluster: kubernetes
    user: myuser
  name: myuser@kubernetes
current-context: myuser@kubernetes
kind: Config
preferences: {}
users:
- name: myuser
  user:
    client-certificate-data: $(cat myuser.crt | base64 -w0)
    client-key-data: $(cat myuser.key | base64 -w0)
EOF
```
### Предоставляем доступ созданному пользователю

Создадим `ClusterAuthorizationRule`:
```yaml
apiVersion: deckhouse.io/v1alpha1
kind: ClusterAuthorizationRule
metadata:
  name: myuser
spec:
  subjects:
  - kind: User
    name: myuser
  accessLevel: PrivilegedUser
  portForwarding: true
```

## Настройка kube-apiserver

Для корректной работы параметра `enableMultiTenancy` необходимо настроить kube-apiserver. Для этого предусмотрен специальный модуль [control-plane-configurator](/modules/160-control-plane-configurator).

{% offtopic title="Изменения манифеста, которые произойдут" %}

* Будет поправлен аргумент `--authorization-mode`, добавится перед методом RBAC метод Webhook (например, --authorization-mode=Node,Webhook,RBAC).
* Добавится `--authorization-webhook-config-file=/etc/kubernetes/authorization-webhook-config.yaml`.
* Добавится `volumeMounts`:

  ```yaml
  - name: authorization-webhook-config
    mountPath: /etc/kubernetes/authorization-webhook-config.yaml
    readOnly: true
  ```
* Добавится `volumes`:

  ```yaml
  - name:authorization-webhook-config
    hostPath:
      path: /etc/kubernetes/authorization-webhook-config.yaml
      type: FileOrCreate
  ```
{% endofftopic %}

## Как проверить, что у пользователя есть доступ?
Необходимо выполнить следующую команду, в которой будут указаны:
* `resourceAttributes` (как в RBAC) - к чему мы проверяем доступ
* `user` - имя пользователя
* `groups` - группы пользователя

P.S. При совместном использовании с модулем `user-authn`, группы и имя пользователя можно посмотреть в логах Dex — `kubectl -n d8-user-authn logs -l app=dex` (видны только при авторизации)
```bash
cat  <<EOF | 2>&1 kubectl create -v=8 -f - | tail -2 \
  | grep "Response Body" | awk -F"Response Body:" '{print $2}' \
  | jq -rc .status
apiVersion: authorization.k8s.io/v1
kind: SubjectAccessReview
spec:
  resourceAttributes:
    namespace: d8-monitoring
    verb: get
    group: ""
    resource: "pods"
  user: "user@gmail.com"
  groups:
  - Everyone
  - Admins
EOF
```
В результате увидим, есть ли доступ и на основании какой роли:
```bash
{
  "allowed": true,
  "reason": "RBAC: allowed by ClusterRoleBinding \"user-authz:myuser:super-admin\" of ClusterRole \"user-authz:super-admin\" to User \"user@gmail.com\""
}
```

## Кастомизация прав для предустановленных accessLevel

Если требуется добавить прав для определённого accessLevel, то достаточно создать ClusterRole с аннотацией `user-authz.deckhouse.io/access-level: <AccessLevel>`. Пример:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  annotations:
    user-authz.deckhouse.io/access-level: PrivilegedUser
  name: d8-mymodule-ns:privileged-user
rules:
- apiGroups:
  - mymodule.io
  resources:
  - destinationrules
  - virtualservices
  - serviceentries
  verbs:
  - create
  - list
  - get
  - update
  - delete
```

## TODO

1. There is a CR `ClusterAuthorizationRule`. Its resources are used to generate `ClusterRoleBindings` for users who mentioned in the field `subjects`. The set of `ClusterRoles` to bind is declared by fields:
    1. `accessLevel` — pre-defined `ClusterRole` set.
    2. `portForwarding` — pre-defined `ClusterRole` set.
    3. `additionalRoles` — user-defined `ClusterRole` set.
2. The configuration of fields `allowAccessToSystemNamespaces` and `limitNamespaces` affects the `user-authz-webhook` DaemonSet, which is authorization agent of apiserver,
3. When creating `ClusterRole` objects with annotation `user-authz.deckhouse.io/access-level`, the set of `ClusterRoles` for binding to the corresponding subject is extended.