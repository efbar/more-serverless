# More Serverless

This repo contains a collection of serverless functions written in GO that can be deployed to services like Openfaas, Google Cloud Functions and Google Cloud Run.

If you want to try OpenFaas locally have a look at [https://github.com/efbar/hashicorp-labs](https://github.com/efbar/hashicorp-labs).

**Table of Contents**
- [More Serverless](#more-serverless)
  - [Usage](#usage)
    - [Requirements](#requirements)
    - [OpenFAAS](#openfaas)
      - [faas up](#faas-up)
      - [faas delete](#faas-delete)
    - [Google Cloud Functions](#google-cloud-functions)
    - [Google Cloud Run](#google-cloud-run)
  - [Functions](#functions)
    - [Google](#google)
      - [gce-toggle](#gce-toggle)
      - [gce-list](#gce-list)
    - [Hashicorp Vault](#hashicorp-vault)
      - [vault-status](#vault-status)
      - [vault-kv-get](#vault-kv-get)
      - [vault-kv-put](#vault-kv-put)
      - [vault-transit](#vault-transit)
    - [Hashicorp Consul](#hashicorp-consul)
      - [consul-catalog-services](#consul-catalog-services)
      - [consul-members](#consul-members)
      - [consul-op-raft-list](#consul-op-raft-list)
    - [Hashicorp Nomad](#hashicorp-nomad)
      - [nomad-job-status](#nomad-job-status)
      - [nomad-node-status](#nomad-node-status)
      - [nomad-server-members](#nomad-server-members)
    - [Slack functions](#slack-functions)
      - [slack-message](#slack-message)

## Usage

Makefile can help you to perform functions building and deploying.

Run the following for some explanation:

```bash
make help
```

### Requirements

For building and deploying automation you need to install:

- docker 
- gcloud
- make
- awk

### OpenFAAS

For OpenFaas you need for sure `faas-cli` and you have to set some variables like:

```bash
export OPENFAAS_URL=http://faasd-gateway:8080
```

You also need to change the image path for every function (needed for docker pushing) in `stack.yml`. You will have to let openfaas login to your image registry correctly. More at OpenFaas documentation [https://docs.openfaas.com](https://docs.openfaas.com)

#### faas up

With this command you will build and deploy to OpenFaas:

```bash
make faasup func=<function_name>
```

#### faas delete

You can delete the function from Openfaas with:

```bash
make faasdelete func=<function_name>
```

### Google Cloud Functions

You can deploy on GCP Cloud Functions once you have setup a project with all the mandatory services enabled (Cloud Functions and Cloud Build for example).

Then you have to choose a function and do:

```bash
make buildgcf func=<function_name> project_id=<project_id> region=<region>
```

where `<function>` is the choosen function, `<project_id>` is the GCP project id and `<region>` is the region for your Cloud Function container.
Optionally can be added some environment variables with `env_vars=<VAR1=value1,VAR2=value2>`.

### Google Cloud Run

The functions can be deployed to Google Cloud Run.

**This automated part needs `faas-cli ` installed.**

Before start, you have to docker login to the GCP registry where the containers will be pull from (us.gcr.io, gcr.io, etc..).

Then:

```bash
make buildgcr func=<function> project_id=<project_id> registry=<registry> region=<region>
```

where `<function>` is the choosen function, `<project_id>` is the GCP project id, `<registry>` is GCP registry where you have just logged in and `<region>` is the region for your Cloud Run container.
Optionally can be added some environment variables with `env_vars=<VAR1=value1,VAR2=value2>`.

## Functions

> Go tested version: v1.16.1

Every folder contains everything to deploy a function. This list will be updated constantly.

### Google

#### gce-toggle

* __description__: stop and start every VM, downscales or scales up (to 3 instances) every managed regional instance group in a GCP project in a "toggle" way
* __input__: project id and region via env variable (look `env_vars`)
* __output__: list of which machine or instance group has been modified
* __env_vars__: in `stack.yml`, under function `environment` key, set `PROJECT_ID` and `REGION` where deploy the function
* __secrets__: in `stack.yml`, under function `secrets` key set `<secret_name>` secret representing the json key file of the service account which has all the permissions you need to call the function (that you have to create with `faas-cli secret create `<secret_name>` --from-file=/path/to/file/sa-key.json`)


#### gce-list

* __description__: same as `gcloud compute instances list` command. Optionally, it can send the output as a message to a Slack Channel.
* __input__: project id and region via env variable (look `env_vars`), you can pass a JSON key file of a service account for authentication and authorization. For sending Slack message `Content-Type` header must be set to `text/plain`. 
* __output__: list every VM in the GCP project defined in `PROJECT_ID`.
* __env_vars__: in `stack.yml`, under function `environment` key, set `PROJECT_ID` and `REGION` where deploy the function (those are mandatory). For sending the output as a Slack message you must add `SLACK_TOKEN` and `SLACK_CHANNEL`, `SLACK_EMOJI` is optional.
* __secrets__: in `stack.yml`, under function `secrets` key set `<secret_name>` secret representing the json key file of the service account which has all the permissions you need to call the function (that you have to create with `faas-cli secret create `<secret_name>` --from-file=/path/to/file/sa-key.json`)

### Hashicorp Vault

#### vault-status

* __description__: same as `vault status` command
* __input__: body: `{"endpoint":"https://vault-endpoint.example"}`
* __output__: same as vault command, content-type could be json and text/plain

#### vault-kv-get

* __description__: same as `vault kv get` command
* __input__: body: `{"token":"s.4w0nd3rfu1t0k3n","endpoint":"https://vault-endpoint.example","path":"secret/data/test","data":{"foo":"bar"}}`, `data` can be empty, `path` needs `data` subpath at the moment.
* __output__: same as vault command, content-type could be json and text/plain

#### vault-kv-put

* __description__: same as `vault kv put` command
* __input__: body: `{"token":"s.4w0nd3rfu1t0k3n","endpoint":"https://vault-endpoint.example","path":"secret/data/test","data":{"foo":"bar"}}`, `data` can not be empty, `path` needs `data` subpath at the moment.
* __output__: same as vault command, content-type could be json and text/plain

#### vault-transit

* __description__: same as `vault transit` command, it can encrypt, decrypt, rewrap, rotate and create new key.
* __input__: body: `{"token":"s.4w0nd3rfu1t0k3n","endpoint":"https://vault-endpoint.example","path":"transit/encrypt/testkey","data":{"plaintext":"Zm9vYmFy"}}`, `data` could be empty only if `path` is not meant for rewrap, rotate or create new key.
* __output__: same as vault command, content-type could be json (in case of encrypt, decrypt and rewrap only) and text/plain

### Hashicorp Consul

#### consul-catalog-services

* __description__: same of `consul catalog services` command
* __input__: body: `{"token":"12345678-1111-2222-3333-a6a53hfd8k1j","endpoint":"https://consul-endpoint.example"}` 
* __output__:  same as consul command but with `-tag` option enabled, content-type could be json and text/plain

#### consul-members 

* __description__: same of `consul members` command  
* __input__: body: `{"token":"12345678-1111-2222-3333-a6a53hfd8k1j","endpoint":"https://consul-endpoint.example"}`
* __output__: same as consul command, content-type could be json and text/plain

#### consul-op-raft-list

* __description__: same as `consul operator raft list-peers` command
* __input__: body: `{"token":"12345678-1111-2222-3333-a6a53hfd8k1j","endpoint":"https://consul-endpoint.example"}`
* __output__: same as consul command, content-type could be json and text/plain

### Hashicorp Nomad

#### nomad-job-status 

* __description__: same as `nomad job status` command
* __input__: body: `{"token":"12345678-1111-2222-3333-a6a53hfd8k1j","endpoint":"https://nomad-endpoint.example"}`
* __output__: same as nomad command, content-type could be json and text/plain

#### nomad-node-status

* __description__: same as `nomad node status` command
* __input__: body: `{"token":"12345678-1111-2222-3333-a6a53hfd8k1j","endpoint":"https://nomad-endpoint.example"}`
* __output__: same as nomad command, content-type could be json and text/plain

#### nomad-server-members

* __description__: same as `nomad server members` command
* __input__: body: `{"token":"12345678-1111-2222-3333-a6a53hfd8k1j","endpoint":"https://nomad-endpoint.example"}`
* __output__: same as nomad command, content-type could be json and text/plain


### Slack functions

#### slack-message

* __description__: send a message to a Slack channel
* __input__: body: `{"token":"xoxp-123456789012-123456789012-123456789012-1234567890121234567890127asd5ff","message":"Hello world","channel":"C123TESTCH1"}`
* __output__: it will logs both message sent positively or not
