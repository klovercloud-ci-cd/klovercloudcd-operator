# klovercloudcd-operator
## Installation
| Available Tags |
|----------------|
| v0.0.1-beta     |
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

#### Create namespace:
```shell
kubectl create ns <name>
```
#### Create Klovercloud CD:
Create a file named ```klovercloudcd.yaml```
```yml
apiVersion: base.cd.klovercloud.com/v1alpha1
kind: KlovercloudCD 
metadata:
 name: klovercloudcd-sample #KlovercloudCD name
 namespace: <your namespace> #namespce where want to install custom resources
spec:
 version: v0.0.1-beta #klovercloudCD version
 db:
   type: MONGO #klovercloudCD database type
   user_name: <database username>
   password: <database password>
   server_url: <database server url>
   server_port: <database server port>
 security:
   user:
     first_name: <user first name>
     last_name: <user last name>
     email: <user email>
     password: <user password>
     phone: <user phone number>
     company_name: <user company name>
#    mail_server_host_email: <user host email>
#    mail_server_host_email_secret: <host email secret>
#    smtp_host: <host email smtp host>
#    smtp_port: <host email smtp port>
   size: 1 #number of replicas
   resources:
     requests:
       cpu: 66m
       memory: 256Mi
     limits:
       cpu: 200m
       memory: 256Mi
 light_house:
   enabled: "true"
   command:
     size: 1
     resources:
       requests:
         cpu: 100m
         memory: 256Mi
       limits:
         cpu: 100m
         memory: 256Mi
   query:
     size: 1
     resources:
       requests:
         cpu: 100m
         memory: 256Mi
       limits:
         cpu: 100m
         memory: 256Mi
 api_service:
   size: 1
   resources:
     requests:
       cpu: 100m
       memory: 256Mi
     limits:
       cpu: 100m
       memory: 256Mi
 agent:
   size: 1
   pull_size: "5" #number of jobs at a time
   light_house_enabled: "true" #to enable lighthouse
   resources:
     requests:
       cpu: 100m
       memory: 256Mi
     limits:
       cpu: 100m
       memory: 256Mi
 integration_manager:
   size: 1
   per_day_total_process: "200" 
   concurrent_process: "10"
   github_webhook_consuming_url: http://<api-service-url>/api/v1/githubs
   bitbucket_webhook_consuming_url: http://<api-service-url>/api/v1/bitbuckets
   pipeline_purging: "ENABLE" #resources purging flag
   resources:
     requests:
       cpu: 100m
       memory: 256Mi
     limits:
       cpu: 100m
       memory: 256Mi
 event_bank:
   size: 1
   resources:
     requests:
       cpu: 100m
       memory: 256Mi
     limits:
       cpu: 100m
       memory: 256Mi
 core_engine:
   size: 1
   number_of_concurrent_process: 5
   resources:
     requests:
       cpu: 100m
       memory: 256Mi
     limits:
       cpu: 100m
       memory: 256Mi

```

#### Apply:
```shell
kubectl apply -f klovercloudcd.yaml
```

#### Create external agent:
Every kubernetes cluster should run an agent; agent needs token to talk to api-service securely.
To generate a token follow the steps:

Exec into api service pod, run:

```
kubectl exec -it <api service pod> -n <namespace> bash
```

Generate agent token, run:

```
kcpctl generate-jwt client=<your agent name> // agent name should be unique
```

It will create a token like,

```
token:  <token>
```

Create a file named ```klovercloud_external_agent.yaml```

```yaml
apiVersion: base.cd.klovercloud.com/v1alpha1
kind: ExternalAgent
metadata:
 name: externalagent-sample
 namespace: klovercloud
spec:
 version: v0.0.1-beta
 agent:
   size: 1
   pull_size: "5"
   light_house_enabled: "true"
   token: <token> #agent token generated from api service
   event_store_url: http://<api-service-url>/api/v1 #api service url
   resources:
     requests:
       cpu: 100m
       memory: 256Mi
     limits:
       cpu: 100m
       memory: 256Mi

```

#### Apply:
```shell
kubectl apply -f klovercloud_external_agent.yaml
```

#### Create UI console:
Create a file named ```klovercloud_ui_console.yaml```
```yaml
apiVersion: base.cd.klovercloud.com/v1alpha1
kind: Console
metadata:
  name: console-sample
  namespace: klovercloud
spec:
  version: v0.0.1-beta
  console:
    size: 1
    auth_endpoint: http://<security-service-endpoint>/api/v1
    api_endpoint: http://<api-service-endpoint>/api/v1
    api_endpoint_ws: ws://<api-service-endpoint>/api/v1
    resources:
      requests:
        cpu: 100m
        memory: 256Mi
      limits:
        cpu: 100m
        memory: 256Mi
```

#### Apply:
```shell
kubectl apply -f klovercloud_ui_console.yaml
```

#### Delete klovercloudCD:
```shell
kubectl delete -f klovercloudcd.yaml
```

#### Delete external agent:
```shell
kubectl delete -f klovercloud_external_agent.yaml
```

#### Delete console:
```shell
kubectl delete -f klovercloud_ui_console.yaml
```

#### Delete operator:
```shell
make undeploy
```
