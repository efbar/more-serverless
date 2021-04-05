PROJECT_NAME=more-serverless
VERSION=1.0.0

## - buildgcf func=<function> project_id=<project_id> region=<region>, deploy function on Google Cloud Function
buildgcf:
    
	@-gcloud config set project $(project_id)
	@-$(eval SUBF := $(shell echo $(func)| tr -d '-'))
	@-cd $(func)/$(SUBF) && go mod vendor && gcloud functions deploy $(func) --entry-point=Serve --runtime=go113 --trigger-http --set-env-vars "PROJECT_ID=$(project_id),REGION=$(region)" --memory 128M --quiet

## - buildgcr func=<function> project_id=<project_id> registry=<registry> region=<region>, build function with OpenFaas builder
buildgcr:

	@-faas-cli build --filter $(func)
	@-$(eval IMAGE := $(shell cat stack.yml| grep vault-read | grep image | awk '{print $$2}'))
	@-$(eval IMAGE_NAME := $(shell echo "$(IMAGE)" | cut -d'/' -f2-))
	@-docker tag $(IMAGE) $(registry)/$(project_id)/$(IMAGE_NAME)
	@-docker push $(registry)/$(project_id)/$(IMAGE_NAME)
	@-gcloud run deploy $(func) --image $(registry)/$(project_id)/$(IMAGE_NAME) --platform managed --memory 128M --region $(region) 

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
