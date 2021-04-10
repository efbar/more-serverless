## $ make buildgcf func=<function> project_id=<project_id> region=<region>
buildgcf:
    
	@-gcloud config set project $(project_id)
	@-$(eval SUBF := $(shell echo $(func)| tr -d '-'))
	@-cd $(func)/$(SUBF) && go mod vendor && gcloud functions deploy $(func) --entry-point=Serve --runtime=go113 --trigger-http --set-env-vars "PROJECT_ID=$(project_id),REGION=$(region)" --memory 128M --quiet

## $ make buildgcr func=<function> project_id=<project_id> registry=<registry> region=<region>
buildgcr:

	@-faas-cli build --filter $(func)
	@-$(eval IMAGE := $(shell cat stack.yml| grep $(func) | grep image | awk '{print $$2}'))
	@-$(eval IMAGE_NAME := $(shell echo "$(IMAGE)" | cut -d'/' -f2-))
	@-docker tag $(IMAGE) $(registry)/$(project_id)/$(IMAGE_NAME)
	@-docker push $(registry)/$(project_id)/$(IMAGE_NAME)
	@-gcloud run deploy $(func) --image $(registry)/$(project_id)/$(IMAGE_NAME) --platform managed --memory 128M --region $(region) 

## $ make faasdelete func=<function>
faasdelete:

	@-faas-cli remove --filter $(func)


## $ make faasup func=<function>
faasup: faasdelete $(func)

	@-faas-cli up --filter $(func)

.PHONY: help
help: Makefile
	@echo
	@echo " 	usage: make <command> <args>"
	@echo " "
	@echo " commands available:"
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ \n\t/'
	@echo
