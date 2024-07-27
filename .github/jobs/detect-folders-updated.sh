#/bin/bash

#### Inputs
# 1 commit before
# 2 current commit
# 3 workflow (infra or app)

## Validate if the workflow is running manually 
CHANGED_FOLDERS=""
echo "Detecting folders updated automatically"
GIT_FOLDERS_UPDATED="$(git diff --name-only $1 $2 )"

add_folder() {
    echo "--> adding the folder $root_folder"
    CHANGED_FOLDERS="$CHANGED_FOLDERS $root_folder"
    echo "--> updating the the variable"
    echo $CHANGED_FOLDERS 
}

for file in $GIT_FOLDERS_UPDATED; do
    folder_name=$(dirname "$file")
    echo "-----> checking the folder $folder_name"
    if [[ "$folder_name" == *"app"* && "$folder_name" == *"-service"* ]] ; then
            echo "Code update for service"
            root_folder=$(echo "$folder_name" | awk -F'/' '{print $2}')
            echo $root_folder
            add_folder        
    elif [[ "$folder_name" == *"4-workloads-stage"* && "$folder_name" == *"-service"* ]] ; then
            echo "Infra update for infra service"
            root_folder=$(echo "$folder_name" | awk -F'/' '{print $3}')
            echo $root_folder
            add_folder   
    elif [[ "$folder_name" == *"-stage"*  ]] ; then
            echo "Infra update for service"
            root_folder=$(echo "$folder_name" | awk -F'/' '{print $2}')
            echo $root_folder
            add_folder 
    else
       echo "----> ignoring the folder $folder_name"
    fi
done

CHANGED_FOLDERS=$(echo "$CHANGED_FOLDERS" | tr ' ' '\n' | sort -u | tr '\n' ' ')

# Create an array for the folders changed to then move to json  
            
read -ra FOLDERS_UPDATED_ARRAY <<< $CHANGED_FOLDERS

# Convert array to JSON structure, this is necessary to export the values as a GH job output

FOLDERS_MODIFIED_JSON="{\"folders\":["
            
for item in "${FOLDERS_UPDATED_ARRAY[@]}"; do
   FOLDERS_MODIFIED_JSON+="\"$item\","
done

FOLDERS_MODIFIED_JSON="${FOLDERS_MODIFIED_JSON%,}"  # Remove the trailing comma
            
FOLDERS_MODIFIED_JSON+="]}"

echo $FOLDERS_MODIFIED_JSON

echo "FOLDERS_UPDATED=$FOLDERS_MODIFIED_JSON" >> $GITHUB_OUTPUT


