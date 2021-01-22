#!/bin/bash
counter=1
while [ $counter -le 18 ]
do
    name="kubectl apply -f cluster"$counter".yaml" 
    echo $name
    $name
    ((counter++))
done
