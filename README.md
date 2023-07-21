# nxs-rbac-operator

## Introduction

nxs-rbac-operator it is a Kubernetes operator creates a role bindings for `groups`, `users` and `service accounts` in a specifed namespaces (by `regex`) within the Kubernetes cluster.

### Features

- Ability to specify the namespaces (by `regex`) for which you want to automate the creation of role bindings
- Ability to use all RBAC subjects (`users`, `groups` and `serviceAcouunts`) for create role bindings
- Ability to use `ClusterRoles` and `Roles` for create role bindings
- Ability to create multiple role bindings for every namespace

### Who use the tool

Development teams and projects who has dynamic namespaces.

## Quickstart

- Clone the repo and go to `.deploy/kubernetes` directory:
  ```
  git clone git@github.com:nixys/nxs-rbac-operator.git
  ```
- Install [Nixys universal Helm chart](https://github.com/nixys/nxs-universal-chart) (`Helm 3` is required):
  ```
  helm repo add nixys https://registry.nixys.ru/chartrepo/public
  ```
- Configure nxs-rbac-operator (see [Configure](#configure) section for details)
- Launch the operator with command:
  ```
  helm -n nxs-rbac-operator --create-namespace install nxs-rbac-operator nixys/universal-chart -f values.yaml
  ```

### Settings

Default configuration file path: `/nxs-rbac-operator.conf`. File represented in yaml.

#### General settings

| Option         | Type   | Required | Default value | Description                                                      |
|---             | :---:  | :---:    | :---:         |---                                                               |
| `logfile`      | String | No       | `stdout`      | Log file path. Also you may use `stdout` and `stderr` |
| `loglevel`     | String | No       | `info`        | Log level. Available values: `debug`, `warn`, `error` and `info` |
| `pidfile`      | String | No       | -             | Pid file path. If `pidfile` is not set it will not be created |
| `kubeConfig`   | String | No       | -             | Path to kubeconfig file. If not set in-cluster kubeconfig will be used |
| `rules`   | List of [Rules](#rules-settings) | Yes       | -             | List of rules to create role bindings for specific namespaces |

##### Rule settings

| Option         | Type   | Required | Default value | Description                                                      |
|---             | :---:  | :---:    | :---:         |---                                                               |
| `ns`      | String | Yes       | -      | Name (regex) of namespaces. For namespaces matched by name with specified regex, rolebindings will be created |
| `roleBindings`      | List of [RoleBindings](#roleBinding-settings) | Yes       | -      | A set of rolebindings to be created in mached namespaces |

##### RoleBinding settings

| Option         | Type   | Required | Default value | Description                                                      |
|---             | :---:  | :---:    | :---:         |---                                                               |
| `role`      | [Role](#role-settings) | Yes       | -      | Role to be bonded with the specified subjects |
| `subjects`      | [Subjects](#subjects-settings) | Yes       | -      | Subjects (`users`, `groups` and `serviceAccounts`) to be bonded with specified role |

##### Role settings

| Option         | Type   | Required | Default value | Description                                                      |
|---             | :---:  | :---:    | :---:         |---                                                               |
| `kind`      | String | Yes       | -      | Role kind (`ClusterRole` or `Role`) to be bonded with the specified subjects |
| `name`      | String | Yes       | -      | Role name to be bonded with the specified subjects |

##### Subjects settings

| Option         | Type   | Required | Default value | Description                                                      |
|---             | :---:  | :---:    | :---:         |---                                                               |
| `users`      | List of strings | No       | -      | User names to be bonded with the role |
| `groups`      | List of strings | No       | -      | Group names to be bonded with the role |
| `serviceAccounts`      | List of strings | No       | -      | Service account names to be bonded with the role |

### Examples

Imaging a case. Every time a new namespaces with the names starts with `dev-` is created a following access need to be granted:
- For the service account `gitlab-deployer`: deploy from CI/CD everything you need for your application within the namespace (cluster role `gitlab-deployer` need to be used for it)
- For the group `admins-l1` (L1 support): view resources such as pod logs and etc. within the namespace (role `view` need to be used for it)
- For the group `admins-l2` (L2 support) and user `localadmin`: full access for all resources within the namespace

The nxs-rbac-operator config file for conditions described above:
```yaml
rules:
- ns: ^dev-.*$
  roleBindings:
  - role:
      kind: ClusterRole
      name: gitlab-deployer
    subjects:
      serviceAccounts:
      - gitlab-deployer
  - role: 
      kind: Role
      name: view
    subjects:
      groups:
      - admins-l1
  - role: 
      kind: Role
      name: admin
    subjects:
      users:
      - localadmin
      groups:
      - admins-l2
```

### Configure

You need to set up the nxs-rbac-operator config file (see options description in [settings section](#settings)). To configure the Operator you need to change the file `.deploy/kubernetes/values.yaml`, secret `nxs-rbac-operator-config` in accordance with your project requirements.

Go back to the [Quickstart section](#quickstart) and follow the instructions to complete the nxs-rbac-operator installation.

## Feedback

For support and feedback please contact me:
- telegram: [@borisershov](https://t.me/borisershov)
- e-mail: b.ershov@nixys.ru

## License

nxs-rbac-operator is released under the [Apache License 2.0](LICENSE).