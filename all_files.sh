#!/bin/bash

# Function to print the structure of a Go project
print_go_project_structure() {
    local directory="$1"
    local output_file="$2"
    
    # Check if the directory exists
    if [ ! -d "$directory" ]; then
        echo "Directory not found: $directory" > "$output_file"
        return
    fi

    # Clear the output file
    > "$output_file"

    # Iterate through all Go files and text files in the directory
    find "$directory" -type f -name "*.go" -o -name "*.txt" | while read -r file; do
        {
            echo "Path: $file"
            echo "Content:"

            # Print the content of the file and remove empty lines
            if [ -r "$file" ]; then
                awk 'NF { print }' "$file" # Remove empty lines
            else
                echo "Could not read file: $file"
            fi

            echo -e "\n----------------------------------------\n"
        } >> "$output_file"
    done
}

# Use current directory
current_dir=$(pwd)

# Specify the path to the Go project directory and the output file
go_project_path="$current_dir/blockchain"
output_file="$current_dir/go_project_structure.txt"

# Call the function
print_go_project_structure "$go_project_path" "$output_file"

echo "Go project structure has been saved to $output_file."