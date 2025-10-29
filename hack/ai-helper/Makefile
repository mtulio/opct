# Create a makefile to build the image of ai-helpers project. I will use the built image to create a custom opct-ai-helper image using ai-helpers as a base image.

BUILD_ARGS ?=
WORKSPACE ?= $(PWD)/.opct-workspace

.PHONY: build-ai-helpers-image
build-ai-helpers-image:
	podman build $(BUILD_ARGS) -f ai-helpers/images/Dockerfile -t ai-helpers:latest ai-helpers/

.PHONY: build-opct-ai-helper-image
build-opct-ai-helper-image: build-ai-helpers-image
	podman build $(BUILD_ARGS) -f Containerfile -t opct-ai-helper:latest .

.PHONY: create-workspace
create-workspace:
	test -d $(WORKSPACE) || mkdir -p $(WORKSPACE)

.PHONY: run-opct-ai-helper
run-opct-ai-helper: build-opct-ai-helper-image create-workspace
	podman run -it \
		-e CLAUDE_CODE_USE_VERTEX=$(CLAUDE_CODE_USE_VERTEX) \
		-e CLOUD_ML_REGION=$(CLOUD_ML_REGION) \
		-e ANTHROPIC_VERTEX_PROJECT_ID=$(ANTHROPIC_VERTEX_PROJECT_ID) \
		-v ~/.config/gcloud:/home/claude/.config/gcloud:ro \
		-v $(WORKSPACE):/workspace \
		-w /workspace \
		opct-ai-helper:latest claude
