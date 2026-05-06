---
name: baseline-debug
description: Diagnose issues with the OPCT baseline index, S3 storage, or CloudFront distribution. Use when baseline results are missing, stale, or CI publish/indexer steps are failing.
allowed-tools: Bash, Read, WebFetch
---

# Baseline Debug

Diagnose issues with the OPCT baseline index, S3 storage, or CloudFront distribution.

## When to use

- Baseline results are missing from `opct adm baseline list`
- Index appears stale or corrupted
- CloudFront serving outdated content
- CI baseline publish/indexer steps failing

## Diagnostic steps

When the user reports baseline issues, follow this sequence:

### 1. Check S3 bucket contents

```bash
# Count total objects
aws s3 ls s3://opct-archive/api/v0/result/summary/ | wc -l

# List objects, save for analysis
aws s3 ls s3://opct-archive/api/v0/result/summary/ > /tmp/bucket-list.txt

# Count by type
grep -c '_latest.json' /tmp/bucket-list.txt   # latest aliases
grep -vc '_latest.json\|index.json' /tmp/bucket-list.txt  # real objects
```

### 2. Download and inspect the index

```bash
aws s3 cp s3://opct-archive/api/v0/result/summary/index.json /tmp/index.json

# Quick analysis
python3 -c "
import json
with open('/tmp/index.json') as f:
    idx = json.load(f)
print(f'Total results: {len(idx[\"results\"])}')
print(f'Latest entries: {len(idx[\"latest\"])}')
print(f'Last update: {idx[\"date\"]}')
# Check for _latest pollution
polluted = [r for r in idx['results'] if '_latest' in r['name']]
print(f'Polluted _latest entries: {len(polluted)}')
# Check version coverage
from collections import Counter
versions = Counter(r.get('openshift_version','?') for r in idx['results'] if '_latest' not in r['name'])
print('Entries per version:')
for v, c in sorted(versions.items()):
    print(f'  {v}: {c}')
"
```

### 3. Compare index vs S3

```bash
python3 -c "
import json
# Load index
with open('/tmp/index.json') as f:
    idx = json.load(f)
indexed_names = {r['name'] for r in idx['results'] if '_latest' not in r['name']}

# Load bucket listing
with open('/tmp/bucket-list.txt') as f:
    lines = f.readlines()
s3_names = set()
for l in lines:
    name = l.strip().split()[-1].replace('.json', '')
    if '_latest' not in name and name != 'index':
        s3_names.add(name)

missing = s3_names - indexed_names
extra = indexed_names - s3_names
print(f'In S3 but not in index: {len(missing)}')
for m in sorted(missing)[:20]:
    print(f'  {m}')
print(f'In index but not in S3: {len(extra)}')
"
```

### 4. Check CloudFront cache

```bash
# Fetch via CloudFront and compare
curl -s https://d23912a6309zf7.cloudfront.net/result/summary/index.json | python3 -c "
import json, sys
idx = json.load(sys.stdin)
print(f'CloudFront index: {len(idx[\"results\"])} results, updated {idx[\"date\"]}')
"

# If stale, invalidate manually
aws cloudfront create-invalidation \
    --distribution-id E3MJR7MT6EHHJC \
    --paths /result/summary/index.json /result/summary/*_latest.json
```

## Common issues

| Symptom | Cause | Fix |
|---------|-------|-----|
| Index has exactly 1000 results | Pagination bug (pre-OPCT-399) | Update to fixed version, rerun indexer |
| `_latest` entries in results array | Filter bug (pre-OPCT-399) | Update to fixed version, rerun indexer (auto-cleans) |
| CloudFront `NoCredentialProviders` | Hardcoded AWS profile (pre-OPCT-399) | Update to fixed version |
| Missing recent versions | Pagination cap or index not rerun | Rerun `opct adm baseline indexer` |
| CloudFront shows old data | Cache not invalidated | Manual invalidation (see step 4) |

## Key constants

- **S3 bucket**: `opct-archive`
- **S3 region**: `us-east-1`
- **CloudFront distribution**: `E3MJR7MT6EHHJC`
- **CloudFront URL**: `https://d23912a6309zf7.cloudfront.net`
- **Index path**: `api/v0/result/summary/index.json`
