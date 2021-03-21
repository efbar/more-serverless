## GO Functions

```bash
gcloud functions deploy trigger --entry-point=Toggle --runtime=go113 --trigger-http --set-env-vars "PROJECT_ID=functest-307416,REGION=us-central1"
```

## OpenFAAS

```bash
export OPENFAAS_URL=http://faasd-gateway:8080
export OPENFAAS_PREFIX=efbar
```
