# klovercloudcd-operator
## Installation

#### Clone:
```shell
git clone https://github.com/klovercloud-ci-cd/klovercloudcd-operator -b <tag>
```
#### Example:
```shell
git clone https://github.com/klovercloud-ci-cd/klovercloudcd-operator -b v0.0.1-beta
```
#### Install:

```sh
make deploy IMG=quay.io/klovercloud/klovercloudcd-operator:<tag>
```
#### Example:
```sh
make deploy IMG=quay.io/klovercloud/klovercloudcd-operator:v0.0.1-beta
```
| Releases  | Documentations                    |
|-------------|-----------------------------------|
| [v0.0.1-beta](https://github.com/klovercloud-ci-cd/klovercloudcd-operator) | [v0.0.1-beta](doc/v0.0.1-beta.md) |

#### Delete operator:
```shell
make undeploy
```
