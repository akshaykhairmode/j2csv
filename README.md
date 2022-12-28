
# j2csv

## Convert JSON to CSV file

Simple tool to convert json files to csv. Removes single line and multi line comments in case of object stream. Sample files [Here](https://github.com/akshaykhairmode/j2csv/tree/main/test-files)

Download Binary from [GitHub](https://github.com/akshaykhairmode/j2csv/tree/main/dist) or build from source.

> go install github.com/akshaykhairmode/j2csv@latest //Needs go 1.19. and above

This will install go binary in your **$GOBIN** (If its set) or at **~/go/bin/j2csv**

### *Example Usage,*

> **$GOBIN/j2csv -f myfile.json** OR **j2csv.exe --f myfile.json**

### *Example Usage for an array of objects,*

> **$GOBIN/j2csv -a -f myfile.json** OR **j2csv.exe --a --f myfile.json**

**Options,**
 

      -a    use this option if its an array of objects. Default type is stream of objects.
      -f string
            --f /home/input.txt (Required)
      -h    Prints command help
      -o string
            --f /home/output.txt
      -uts string
            used to convert timestamp to string, usage --uts createdAt,updatedAt
      -v    Enables verbose logging
