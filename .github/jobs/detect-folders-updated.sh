#!/bin/bash
set -euo pipefail  # Exit on error, undefined variables, and pipe failures

#### Inputs
# 1: commit before (required)
# 2: current commit (required)
# 3: scope - 'infra', 'app', or 'all' (optional, defaults to 'all')

# Validate and set input parameters
COMMIT_BEFORE="${1:?Error: Missing required parameter - commit before}"
COMMIT_CURRENT="${2:?Error: Missing required parameter - current commit}"
SCOPE="${3:-all}"

echo "============================================"
echo "Detecting folders updated automatically"
echo "  Commit range: $COMMIT_BEFORE...$COMMIT_CURRENT"
echo "  Scope: $SCOPE"
echo "============================================"

# Validate scope parameter
if [[ ! "$SCOPE" =~ ^(infra|app|all)$ ]]; then
    echo "Error: Invalid scope '$SCOPE'. Must be 'infra', 'app', or 'all'"
    exit 1
fi

# Get changed files with error handling
if ! GIT_FOLDERS_UPDATED="$(git diff --name-only "$COMMIT_BEFORE" "$COMMIT_CURRENT" 2>&1)"; then
    echo "Error: git diff failed"
    echo "$GIT_FOLDERS_UPDATED"
    exit 1
fi

# Check if there are any changes
if [[ -z "$GIT_FOLDERS_UPDATED" ]]; then
    echo "No files changed between commits"
    echo "FOLDERS_UPDATED={\"folders\":[]}" >> "$GITHUB_OUTPUT"
    exit 0
fi

# Use associative array for automatic deduplication
declare -A CHANGED_FOLDERS_MAP

# Process each changed file
for file in $GIT_FOLDERS_UPDATED; do
    folder_name=$(dirname "$file")
    echo "Checking: $folder_name"

    # Check for app services (e.g., app/user-service/*)
    if [[ ("$SCOPE" == "app" || "$SCOPE" == "all") && "$folder_name" =~ ^app/[^/]*-service(/|$) ]]; then
        root_folder=$(echo "$folder_name" | cut -d'/' -f2)
        echo "  ✓ App service detected: $root_folder"
        CHANGED_FOLDERS_MAP["$root_folder"]=1

    # Check for infra workload services (e.g., infra/4-workloads-stage/user-service/*)
    elif [[ ("$SCOPE" == "infra" || "$SCOPE" == "all") && "$folder_name" =~ ^infra/4-workloads-stage/[^/]*-service(/|$) ]]; then
        root_folder=$(echo "$folder_name" | cut -d'/' -f3)
        echo "  ✓ Infra workload service detected: $root_folder"
        CHANGED_FOLDERS_MAP["$root_folder"]=1

    # Check for infra stages (e.g., infra/1-base-stage/*, infra/2-eks-cluster-stage/*)
    elif [[ ("$SCOPE" == "infra" || "$SCOPE" == "all") && "$folder_name" =~ ^infra/[0-9]+-[^/]*-stage(/|$) ]]; then
        root_folder=$(echo "$folder_name" | cut -d'/' -f2)
        echo "  ✓ Infra stage detected: $root_folder"
        CHANGED_FOLDERS_MAP["$root_folder"]=1
    else
        echo "  ✗ Ignored (out of scope): $folder_name"
    fi
done

# Convert associative array keys to sorted JSON array
if [[ ${#CHANGED_FOLDERS_MAP[@]} -eq 0 ]]; then
    echo ""
    echo "No relevant folders detected for scope: $SCOPE"
    FOLDERS_MODIFIED_JSON='{"folders":[]}'
else
    echo ""
    echo "Detected folders:"
    printf '  - %s\n' "${!CHANGED_FOLDERS_MAP[@]}" | sort

    # Convert to JSON using jq for proper escaping
    FOLDERS_MODIFIED_JSON=$(printf '%s\n' "${!CHANGED_FOLDERS_MAP[@]}" | sort | jq -R . | jq -s -c '{folders: .}')
fi

echo ""
echo "Final output: $FOLDERS_MODIFIED_JSON"
echo "FOLDERS_UPDATED=$FOLDERS_MODIFIED_JSON" >> "$GITHUB_OUTPUT"


