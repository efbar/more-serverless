# More Serverless

A collection of serverless functions written in GO.

**Table of Contents**
- [More Serverless](#more-serverless)
  - [Usage](#usage)
    - [Google Cloud Functions](#google-cloud-functions)
    - [OpenFAAS](#openfaas)

## Usage

This repo contains functions to be deployed with services like Google Cloud Functions and similar.
But it provides also some components to let you deploy them in OpenFaas.

If you want to try OpenFaas locally have a look at [https://github.com/efbar/hashicorp-labs](https://github.com/efbar/hashicorp-labs).

Every folder contains everything to deploy a function. The provided functions are (this list will be updated constantly):

|  Function | Description |
| --- | --- |
|gce-toggle|  |
|consul-catalog-services|  |
|consul-members|  |
|consul-op-raft-list|  |
|nomad-job-status|  |
|nomad-node-status|  |
|nomad-server-members|  |

In the root folder you can see `stack.yml`. This file is useful to deploy function in OpenFaas, read above for a quick guide.

> Go tested version: v1.16.1

### Google Cloud Functions

You can deploy on GCP Cloud Functions once you have setup a project with all the mandatory services enabled (Cloud Functions and Cloud Build for example).
Obviously you need `gcloud` tool and `go` (v1.16.1) installed.

then we have to choose a function and do:

```bash
make buildgcf func=consul-members project_id=functest-307416 region=us-central1
```

then wait for building and deploying.

### OpenFAAS

For OpenFaas you need for sure `faas-cli` and you have to set some variables like:

```bash
export OPENFAAS_URL=http://faasd-gateway:8080
```

You also need to change the image path for every function (needed for docker pushing) in `stack.yml`. You will have to let openfaas login to your image registry correctly. More at OpenFaas documentation [https://docs.openfaas.com](https://docs.openfaas.com)

Then simply:

```bash
make faasup func=consul-members
```
