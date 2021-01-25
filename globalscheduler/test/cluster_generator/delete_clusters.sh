/*
Copyright 2020 Authors of Arktos.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

#!/bin/bash
counter=1
while [ $counter -le 9 ]
do
    name="kubectl delete cluster cluster00$counter"
    echo $name
    $name
     ((counter++))
done
counter=10
while [ $counter -le 99 ]
do
    name="kubectl delete cluster cluster0$counter"
    echo $name
    $name
    ((counter++))
done
counter=100
while [ $counter -le 999 ]
do
    name="kubectl delete cluster cluster$counter"
    $name
    echo $name
    ((counter++))
done
kubectl delete cluster cluster1000