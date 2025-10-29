
# OPCT AI Helper


## Build Image

```sh
make build-opct-ai-helper-image
```


## Usage

### AI Helpers

```sh

```

### OPCT AI Helpers

```sh
# Create the workspace for your review environment
export WORKSPACE=.opct-workspace
mkdir $WORKSPACE


# Run it
podman run -it \
   --userns=keep-id \
  -e CLAUDE_CODE_USE_VERTEX=$CLAUDE_CODE_USE_VERTEX \
  -e CLOUD_ML_REGION=$CLOUD_ML_REGION \
  -e ANTHROPIC_VERTEX_PROJECT_ID=$ANTHROPIC_VERTEX_PROJECT_ID \
  -v ~/.config/gcloud:/home/claude/.config/gcloud:Z \
  -v $WORKSPACE:/workspace:Z \
  -w /workspace \
  opct-ai-helper:latest claude
```
