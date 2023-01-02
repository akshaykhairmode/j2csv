

# j2csv

## Convert JSON to CSV file

Simple tool to convert json files to csv. Removes single line and multi line comments in case of object stream. Sample files [Here](https://github.com/akshaykhairmode/j2csv/tree/main/test-files)

Download Binary from [GitHub](https://github.com/akshaykhairmode/j2csv/tree/main/dist) or build from source.

> go install github.com/akshaykhairmode/j2csv@latest //Needs go 1.19. and above

This will install go binary in your **$GOBIN** (If its set) or at **~/go/bin/j2csv**

**Options available**

      -a    use this option if its an array of objects
      -f string
            usage --f /home/input.txt (Required)
      -force
            force load input file in memory, use this if conversion is failing.
      -h    Prints command help
      -i    get input data from standard input
      -o string
            usage --o /home/output.txt
      -stats
            prints the allocations at start and at end
      -uts string
            used to convert timestamp to string, usage --uts createdAt,updatedAt
      -v    Enables verbose logging
      -z    output file to be .zip



### *Examples,*

#### Normal stream

    ./dist/linux64/j2csv  -f  test-files/object.txt
    
    //Output
    10:26PM INF Reading input from path : test-files/object.txt
    10:26PM INF Output File ====> j2csv-object-1672678561.csv
    10:26PM INF Done!!, Time took : 44.8119ms

#### JSON array stream

    ./dist/linux64/j2csv -a -f test-files/array.json
    
    //Output
    10:28PM INF Reading input from path : test-files/array.json
    10:28PM INF Output File ====> j2csv-array-1672678733.csv
    10:28PM INF Done!!, Time took : 32.2386ms

#### Standard Input

    echo -n '{"key1":"value1","key2":"value2"}' | ./dist/linux64/j2csv -i
    
    //Output
    10:34PM INF Output File ====> j2csv-stdin-1672679072.csv
    10:34PM INF Done!!, Time took : 341.8µs

#### With Force flag, loads the input file in memory instead of streaming. Should be used when conversion is failing

    ./dist/linux64/j2csv -f test-files/object_fail.txt -force
    
    //Output
    10:35PM INF Reading input from path : test-files/object_fail.txt
    10:35PM INF Output File ====> j2csv-object_fail-1672679154.csv
    10:35PM INF Done!!, Time took : 70.8357ms

#### Zip Output

zip input is by default supported **(Only works with single file in zip)**. For zip output use -z.

    ./dist/linux64/j2csv -z -f test-files/object.zip
    
    //Output
    10:38PM INF Reading input from path : test-files/object.zip
    10:38PM INF Output File ====> j2csv-object-1672679327.zip
    10:38PM INF Done!!, Time took : 37.0841ms

#### With custom output path

    ./dist/linux64/j2csv -f test-files/object.zip -o myfile.csv
    
    //Output
    10:43PM INF Reading input from path : test-files/object.zip
    10:43PM INF Output File ====> myfile.csv
    10:43PM INF Done!!, Time took : 40.0745ms

#### Converting unix timestamp to string

    ./dist/linux64/j2csv -f test-files/object.zip -uts createdAt,updatedAt
    
    //Output
    10:44PM INF Reading input from path : test-files/object.zip
    10:44PM INF Output File ====> j2csv-object-1672679695.csv
    10:44PM INF Done!!, Time took : 44.315ms
