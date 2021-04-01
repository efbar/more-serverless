PROJECT_NAME=more-serverless
VERSION=1.0.0

## - buildgcf func=<function> project_id=<project_id> region=<region>, deploy function on Google Cloud Function
buildgcf:
    
	@-gcloud config set project $(project_id)
	@-$(eval SUBF := $(shell echo $(func)| tr -d '-'))
	@-cd $(func)/$(SUBF) && gcloud functions deploy $(func) --entry-point=Serve --runtime=go113 --trigger-http --set-env-vars "PROJECT_ID=$(project_id),REGION=$(region)" --memory 128M --quiet

## - faasdelete func=<function>, delete function from OpenFaas
faasdelete:

	@-faas-cli remove --filter $(func)


## - faasup func=<function>, deploy function on OpenFaas
faasup: faasdelete $(func)

	@-faas-cli up --filter $(func)

.PHONY: help test clean
help: Makefile
	@echo
	@echo " Choose a command from list below"
	@echo " usage: make <command>"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo
